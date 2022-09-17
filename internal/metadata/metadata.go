package metadata

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/toaster/tmpc/internal/mpd"
)

// CoverID returns a unique ID to identify the cover of a song.
func CoverID(song *mpd.Song) string {
	if song.MBAlbumID != "" {
		return song.MBAlbumID
	}
	return SongID(song)
}

// ExtractLyricsFromHTML is a helper to extract lyrics from an HTML page.
func ExtractLyricsFromHTML(nodes []*html.Node, excludeParams map[string][]Matcher) []string {
	var lines []string
	var buf strings.Builder
	appendToBuf := func(s string) {
		if s == "" {
			return
		}

		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(s)
	}
	appendLine := func(line string) {
		lines = append(lines, line)
		buf.Reset()
	}
	appendBufLine := func(force bool) {
		if force || buf.Len() > 0 {
			appendLine(buf.String())
		}
	}
	appendSubLines := func(subLines []string) {
		lines = append(lines, subLines...)
	}
	for _, node := range nodes {
		if len(lines) > 0 {
			appendLine("")
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.TextNode:
				appendToBuf(strings.TrimSpace(c.Data))
			case html.ElementNode:
				if NodeParamsMatch(c, excludeParams) {
					continue
				}

				switch c.DataAtom {
				case atom.Br:
					appendBufLine(true)
				case atom.P, atom.Div:
					appendBufLine(false)
					appendLine("")
					appendSubLines(ExtractLyricsFromHTML([]*html.Node{c}, nil))
					appendLine("")
				default:
					subLines := ExtractLyricsFromHTML([]*html.Node{c}, nil)
					if len(subLines) > 0 {
						if len(subLines) > 1 {
							appendSubLines(subLines[:len(subLines)-1])
						}
						appendToBuf(subLines[len(subLines)-1])
					}
				}
			}
		}
		appendBufLine(false)
	}
	if len(lines) > 0 {
		for line := lines[0]; line == "" && len(lines) > 0; {
			lines = lines[1:]
			if len(lines) > 0 {
				line = lines[0]
			}
		}
		if len(lines) > 0 {
			for line := lines[len(lines)-1]; line == "" && len(lines) > 0; {
				lines = lines[:len(lines)-1]
				if len(lines) > 0 {
					line = lines[len(lines)-1]
				}
			}
		}
	}
	return lines
}

// NewExactMatcher returns a new matcher for matching the exact value.
func NewExactMatcher(value string) Matcher {
	return &matcher{
		prefix: false,
		value:  value,
	}
}

// NewPrefixMatcher returns a new matcher for matching the prefix value.
func NewPrefixMatcher(value string) Matcher {
	return &matcher{
		prefix: true,
		value:  value,
	}
}

// NodeParamsMatch returns whether the attributes of the given html.Node match one of the given matchers.
// The actual value of an attribute is split by space and every single component is tested against the matchers for this attribute.
// The method returns `true` as soon as any value matches.
func NodeParamsMatch(node *html.Node, matchers map[string][]Matcher) bool {
	for _, a := range node.Attr {
		for _, m := range matchers[a.Key] {
			for _, v := range strings.Split(a.Val, " ") {
				if m.MatchPrefix() {
					if strings.HasPrefix(v, m.Value()) {
						return true
					}
				} else {
					if v == m.Value() {
						return true
					}
				}
			}
		}
	}
	return false
}

// ReducedTitle tries to convert and shorten a title to a minimal common part.
// This might help match titles with different writings.
func ReducedTitle(title string, language string) string {
	reduced, _, _ := strings.Cut(title, ":")
	reduced, _, _ = strings.Cut(title, ",")
	reduced = generalTitleReplacer.Replace(reduced)
	reduced = titleReplacers[language].Replace(reduced)
	reduced = replaceNums(reduced, language)
	reduced = reduceRegexp.ReplaceAllString(reduced, "")
	return strings.TrimSpace(reduced)
}

// SongID returns a unique ID to identify a song.
func SongID(song *mpd.Song) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(song.File)))
}

type Matcher interface {
	MatchPrefix() bool
	Value() string
}

var generalTitleReplacer *strings.Replacer
var numRegexp *regexp.Regexp
var numReplacements map[string]map[string]string
var reduceRegexp *regexp.Regexp
var titleReplacers map[string]*strings.Replacer

func init() {
	generalTitleReplacer = strings.NewReplacer("‐", "", "-", "", "  ", " ", "'", "", "’", "", "‘", "")
	numRegexp = regexp.MustCompile("\\b\\d+\\b")
	numReplacements = map[string]map[string]string{
		"de": {"0": "null", "1": "eins", "2": "zwei", "3": "drei", "4": "vier", "5": "fuenf", "6": "sechs", "7": "sieben", "8": "acht", "9": "neun", "10": "zehn", "11": "elf", "12": "zwoelf"},
		"en": {"0": "zero", "1": "one", "2": "two", "3": "three", "4": "four", "5": "five", "6": "six", "7": "seven", "8": "eight", "9": "nine", "10": "ten", "11": "eleven", "12": "twelve"},
	}
	reduceRegexp = regexp.MustCompile("\\(.*\\)")
	titleReplacers = map[string]*strings.Replacer{
		"de": strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss", "&", "und"),
		"en": strings.NewReplacer("&", "and"),
	}
}

func replaceNums(in, lang string) string {
	return numRegexp.ReplaceAllStringFunc(in, func(match string) string {
		replacement := numReplacements[lang][match]
		if replacement == "" {
			replacement = in
		}
		return replacement
	})
}

type matcher struct {
	prefix bool
	value  string
}

var _ Matcher = (*matcher)(nil)

func (m *matcher) MatchPrefix() bool {
	return m.prefix
}

func (m *matcher) Value() string {
	return m.value
}
