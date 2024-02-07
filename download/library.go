package download

import (
	"encoding/json"
	"gotube/config"
	"gotube/download/network"
	"gotube/youtube"
	"strconv"
	"strings"
)

func GetLibrary() youtube.VideoHolder {
	config.LogEvent("Getting library")
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com/feed/library", true)
	config.FileDump("LibraryRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, false)
	config.FileDump("LibraryRaw.json", jsonText, false)
	// Format into correct structure
	var jsonA LibraryJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}

	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("LibraryProcessed.json", string(text), false)
	contents := jsonA.Contents.TwoColumnBrowseResultsRenderer.Tabs[0]
	contentsB := contents.TabRenderer.Content.SectionListRenderer.Contents
	contentsA := contentsB[2].ItemSectionRenderer.Contents[0].ShelfRenderer.Content.GridRenderer.Items
	playlists := []youtube.Video{}

	var doneChan chan int = make(chan int)
	var err error
	_ = err
	var number int = 0
	for _, x := range contentsA {

		playlistJSON := x.GridPlaylistRenderer
		if playlistJSON.Title.SimpleText != "" {

			// Last Updated
			var lastUpdated string = "Unknown"
			if playlistJSON.PublishedTimeText.SimpleText != "" {
				lastUpdated = playlistJSON.PublishedTimeText.SimpleText
				if strings.Contains(lastUpdated, "yesterday") {
					lastUpdated = "Yesterday"
				} else if strings.Contains(lastUpdated, "today") {
					lastUpdated = "Today"
				} else if strings.Contains(lastUpdated, "days ago") {

				} else if strings.Contains(lastUpdated, "months ago") {

				} else if strings.Contains(lastUpdated, "years ago") {

				}
			}

			// Num Videos
			var numVideos int = 0
			if playlistJSON.VideoCountText.Runs != nil {
				//Print(playlistJSON.Title.SimpleText + ": " + playlistJSON.VideoCountText.Runs[0].Text)
				numVideos, err = strconv.Atoi(playlistJSON.VideoCountText.Runs[0].Text)
			}

			var visibility string = "Unknown"

			var author string = "Unknown"
			if playlistJSON.ShortBylineText.Runs[0].NavigationEndpoint.ClickTrackingParams != "" {
				author = playlistJSON.ShortBylineText.Runs[0].Text
				visibility = "Public"
			} else {
				visibility = playlistJSON.ShortBylineText.Runs[0].Text
			}

			number++
			var typeA int = youtube.OTHER_PLAYLIST
			if author == "Unknown" {
				typeA = youtube.MY_PLAYLIST
			}

			// Put it all together
			playlist := youtube.Video{
				Title:         playlistJSON.Title.SimpleText,
				LastUpdated:   lastUpdated,
				NumVideos:     numVideos,
				Channel:       author,
				Visibility:    visibility,
				Id:            playlistJSON.PlaylistID,
				ThumbnailLink: playlistJSON.Thumbnail.Thumbnails[0].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				Type:          typeA,
			}
			playlists = append(playlists, playlist)
			go network.DownloadThumbnail(playlist.ThumbnailLink, playlist.ThumbnailFile, false, doneChan, false)
		}
	}
	for i := 0; i < number; i++ {
		_ = <-doneChan
	}

	holder := youtube.VideoHolder{
		Videos:            playlists,
		PageType:          youtube.LIBRARY,
		ContinuationToken: "",
	}

	return holder
}
