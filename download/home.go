package download

import (
	"gotube/youtube"
	"gotube/config"
	"encoding/json"
	"strconv"
	"strings"
	"gotube/download/network"
)

func GetRecommendations() youtube.VideoHolder {
	config.LogEvent("Getting home page")
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com", true)
	config.FileDump("HomeRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, false)
	config.FileDump("HomeRaw.json", jsonText, false)
	
	// Format into correct structure
	var jsonA RecJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("HomeProcessed.json", string(text), false)
	
	contents := jsonA.Contents.TwoColumnBrowseResultsRenderer.Tabs[0]
	contentsA := contents.TabRenderer.Content.RichGridRenderer.Contents
	videos := []youtube.Video{}
	
	
	var doneChan chan int = make(chan int)
	var err error
	_ = err
	var number int = 0
	for _, x := range contentsA {
		
		videoJSON := x.RichItemRenderer.Content.VideoRenderer
		if videoJSON.Title.Runs != nil {
			
			// Views
			/*
			var views int = 0
			if videoJSON.ViewCountText.Runs == nil {
				simpleText := videoJSON.ViewCountText.SimpleText
				if index := strings.Index(simpleText, " "); index != -1 {
					views, err = strconv.Atoi(strings.ReplaceAll(simpleText[:index], ",", ""))
				}
			} else {
				views, err = strconv.Atoi(strings.ReplaceAll(videoJSON.ViewCountText.Runs[0].Text, ",", ""))
			}
			*/
			var views string
			var vidType string
			if videoJSON.ShortViewCountText.Runs == nil {
				 views = strings.Split(videoJSON.ShortViewCountText.SimpleText, " ")[0]
				 vidType = "Video"
			} else {
				views = videoJSON.ShortViewCountText.Runs[0].Text
				vidType = "Livestream"
			}
			//views = videoJSON.ShortViewCountText.SimpleText
			
			// Published Time
			var releaseDate string = "Livestream"
			if videoJSON.PublishedTimeText.SimpleText != "" {
				releaseDate = videoJSON.PublishedTimeText.SimpleText
			}
			
			// Length
			var length string = "Livestream"
			if videoJSON.LengthText.SimpleText != "" {
				length = videoJSON.LengthText.SimpleText
			}
			
			number++
			_ = views
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
				ThumbnailLink: videoJSON.Thumbnail.Thumbnails[0].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				DirectLink: "",
				StartTime: videoJSON.NavigationEndpoint.WatchEndpoint.StartTimeSeconds,
				Type: youtube.VIDEO,
			}
			videos = append(videos, video)
			go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, false)
		}
	}
	
	videoHolder := youtube.VideoHolder {
		Videos: videos,
		PageType: youtube.HOME,
		ContinuationToken: "",
	}	
	
	for i:=0; i<number; i++ {
		_ = <- doneChan
	}
	return videoHolder
}
