package mpd

import (
	"fmt"
	"strconv"

	"github.com/fhs/gompd/mpd"
)

/*
Notes on MPD data:

Song details: {
	"Album",
	"AlbumArtist",
	"AlbumArtistSort",
	"Artist",
	"ArtistSort",
	"Date"
	"Disc",
	"duration",
	"file", <-- the only mandatory
	"Genre",
	"Id", <-- if in queue
	"Label",
	"Last-Modified",
	"MUSICBRAINZ_ALBUMARTISTID",
	"MUSICBRAINZ_ALBUMID",
	"MUSICBRAINZ_ARTISTID",
	"MUSICBRAINZ_TRACKID",
	"OriginalDate",
	"Time",
	"Title",
	"Track",
}

Ideas:

- types for artists?
	-> only useful with a real (in memory) database to relate data
	-> like “other songs of his artist”, “albums of this artist”, etc.

- use musicbrainz infos for determining album and grouping by album
	-> utility methods here instead of UI

*/

// Song represents a song on the MPD server
type Song struct {
	File        string // mandatory
	Album       string
	AlbumArtist string
	Artist      string
	ID          int
	MBAlbumID   string
	Time        int
	Title       string
	Track       int
	Year        int
}

// DisplayTitle returns the display title of the song.
// It contains the artist iff it differs from the album artist.
func (s *Song) DisplayTitle() string {
	if s.Artist != s.AlbumArtist {
		return fmt.Sprintf("%s - %s", s.Artist, s.Title)
	}
	return s.Title
}

func songsFromAttrs(attrs []mpd.Attrs) []*Song {
	songs := make([]*Song, len(attrs))
	for i, sAttrs := range attrs {
		id, _ := strconv.Atoi(sAttrs["Id"])
		time, _ := strconv.Atoi(sAttrs["Time"])
		if time == 0 {
			time, _ = strconv.Atoi(sAttrs["duration"])
		}
		track, _ := strconv.Atoi(sAttrs["Track"])
		var year int
		fmt.Sscanf(sAttrs["OriginalDate"], "%d", &year)
		if year == 0 {
			fmt.Sscanf(sAttrs["Date"], "%d", &year)
		}
		songs[i] = &Song{
			File:        sAttrs["file"],
			Album:       sAttrs["Album"],
			AlbumArtist: sAttrs["AlbumArtist"],
			Artist:      sAttrs["Artist"],
			ID:          id,
			MBAlbumID:   sAttrs["MUSICBRAINZ_ALBUMID"],
			Time:        time,
			Title:       sAttrs["Title"],
			Track:       track,
			Year:        year,
		}
	}
	return songs
}
