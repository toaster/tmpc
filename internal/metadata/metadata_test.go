package metadata_test

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/toaster/tmpc/internal/metadata"
)

//go:embed testdata
var testdata embed.FS

func TestExtractLyricsFromHTML(t *testing.T) {
	for name, tt := range map[string]struct {
		nodes []*html.Node
		want  []string
	}{
		"plain text": {
			nodes: []*html.Node{
				{
					FirstChild: &html.Node{
						Type: html.TextNode,
						Data: " some line of plain text\t\n",
						NextSibling: &html.Node{
							Type: html.TextNode,
							Data: "line 2",
							NextSibling: &html.Node{
								Type: html.TextNode,
								Data: " more more more! ",
								NextSibling: &html.Node{
									Type: html.TextNode,
									Data: "four",
									NextSibling: &html.Node{
										Type: html.TextNode,
										Data: "END!\n\n",
									},
								},
							},
						},
					},
				},
				{
					FirstChild: &html.Node{
						Type: html.TextNode,
						Data: " second bunch of stuff :D\t\n",
					},
				},
			},
			want: []string{"some line of plain text line 2 more more more! four END!", "", "second bunch of stuff :D"},
		},
		"HTML with divs, paragraphs, links and styles": {
			nodes: parseHTMLFragments(t, "testdata/famous_song_lyrics_fragments.html"),
			want: []string{
				"This is the song",
				"You’re longing for",
				"",
				"Famous indeed",
				"It is all made up",
				"Hell, yeah!",
				"Famous Song!",
				"",
				"This song ain’t over",
				"",
				"We won’t stop playing for you",
				"",
				"At least, not until now",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := metadata.ExtractLyricsFromHTML(tt.nodes, map[string]string{
				"exclude-filter-attr": "exclude-me",
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

func parseHTMLFragments(t *testing.T, filename string) []*html.Node {
	f, err := testdata.Open(filename)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	rawNodes, err := html.ParseFragment(f, &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     atom.Body.String(),
	})
	require.NoError(t, err)

	var nodes []*html.Node
	for _, node := range rawNodes {
		if node.Type == html.ElementNode {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
