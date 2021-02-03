package metadata

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/toaster/tmpc/internal/mpd"
)

// CoverID returns a unique ID to identify the cover of a song.
func CoverID(song *mpd.Song) string {
	if song.MBAlbumID != "" {
		return song.MBAlbumID
	}
	return SongID(song)
}

// ReducedTitle tries to convert and shorten a title to a minimal common part.
// This might help match titles with different writings.
func ReducedTitle(title string, language string) string {
	reduced := strings.SplitN(title, ":", 2)[0]
	reduced = strings.SplitN(title, ",", 2)[0]
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
