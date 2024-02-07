package download

import (
	"gotube/youtube"
	"gotube/config"
	"encoding/json"
	"strconv"
	"strings"
	"os"
	"gotube/download/network"
)

func GetSearch(searchTerm string) youtube.VideoHolder {
	config.LogEvent("Getting search: " + searchTerm)
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com/results?search_query=" + strings.ReplaceAll(searchTerm, " ", "+"), true)
	config.FileDump("SearchRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, false)
	config.FileDump("SearchRaw.json", jsonText, false)
	// Format into correct structure
	var jsonA SearchJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("SearchProcessed.json", string(text), false)
	
	contents := jsonA.Contents.TwoColumnSearchResultsRenderer.PrimaryContents
	contentsB := contents.SectionListRenderer.Contents
	contentsA := contentsB[0].ItemSectionRenderer.Contents
	if 1 > 2 {
		os.Exit(1)
	}
	
	contentsA = append(contentsA, contentsB[1].ItemSectionRenderer.Contents...)
	videos := []youtube.Video{}
	
	var doneChan chan int = make(chan int)
	var err error
	_ = err
	var number int = 0
	for _, x := range contentsA {
		
		videoJSON := x.VideoRenderer
		playlistJSON := x.PlaylistRenderer
		if videoJSON.Title.Runs != nil {
			// Views
			
			var views string = ""
			var vidType string = ""
			if videoJSON.ShortViewCountText.Runs == nil {
				 views = strings.Split(videoJSON.ShortViewCountText.SimpleText, " ")[0]
				 vidType = "Video"
			} else {
				views = videoJSON.ShortViewCountText.Runs[0].Text
				vidType = "Livestream"
			}
			
			// Published Time
			var releaseDate string = "Unknown"
			if videoJSON.PublishedTimeText.SimpleText != "" {
				releaseDate = videoJSON.PublishedTimeText.SimpleText
			}
			
			// Length
			var length string = "Livestream"
			if videoJSON.LengthText.SimpleText != "" {
				length = videoJSON.LengthText.SimpleText
			}
			_ = views
			number++
			// Put it all together
			video := youtube.Video{
				Title: videoJSON.Title.Runs[0].Text,
				Views: views,
				VidType: vidType,
				ReleaseDate: releaseDate,
				Length: length,
				Id: videoJSON.VideoID,
				Channel: videoJSON.OwnerText.Runs[0].Text,
				ChannelID: videoJSON.OwnerText.Runs[0].NavigationEndpoint.CommandMetadata.WebCommandMetadata.URL,
				ThumbnailLink: videoJSON.Thumbnail.Thumbnails[len(videoJSON.Thumbnail.Thumbnails)-1].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				DirectLink: "",
				StartTime: videoJSON.NavigationEndpoint.WatchEndpoint.StartTimeSeconds,
				Type: youtube.VIDEO,
			}
			videos = append(videos, video)
			go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, false)
		} else if playlistJSON.Thumbnails != nil {
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
			// Put it all together
			playlist := youtube.Video{
				Title: playlistJSON.Title.SimpleText,
				LastUpdated: lastUpdated,
				NumVideos: numVideos,
				Channel: author,
				Visibility: visibility,
				Id: playlistJSON.PlaylistID,
				ThumbnailLink: playlistJSON.Thumbnails[0].Thumbnails[0].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				Type: youtube.OTHER_PLAYLIST,
			}
			videos = append(videos, playlist)
			go network.DownloadThumbnail(playlist.ThumbnailLink, playlist.ThumbnailFile, false, doneChan, false)
		}
	}
	
	videoHolder := youtube.VideoHolder {
		Videos: videos,
		PageType: youtube.SEARCH,
		ContinuationToken: "",
	}
	
	for i:=0; i<number; i++ {
		_ = <- doneChan
	}
	
	return videoHolder
}
