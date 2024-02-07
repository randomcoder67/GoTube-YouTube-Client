package download

import (
	"gotube/youtube"
	"gotube/config"
	"encoding/json"
	"strconv"
	"strings"
	"fmt"
	"os"
	"gotube/download/network"
)

// Every page type (search, playlist, history etc) has to be a seperate function as the JSON for each is subtly different. It's annoying but there's no real way to combine these functions

const THUMBNAIL_DIR string = "/.cache/gotube/thumbnails/"

// This file contains various functions to extract a list of videos from your subscriptions, history, watch later, and a search, as well as a list of playlists from your library. 

func GetSubscriptions() youtube.VideoHolder {
	config.LogEvent("Getting subscriptions")
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com/feed/subscriptions", true)
	config.FileDump("SubscriptionsRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, false)
	config.FileDump("SubscriptionsRaw.json", jsonText, false)
	// Format into correct structure
	var jsonA SubJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("SubscriptionsProcessed.json", string(text), false)
	
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
				ThumbnailLink: videoJSON.Thumbnail.Thumbnails[2].URL,
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
		PageType: youtube.SUBS,
		ContinuationToken: "",
	}
	
	for i:=0; i<number; i++ {
		_ = <- doneChan
	}
	return videoHolder
}

func GetHistory() youtube.VideoHolder {
	config.LogEvent("Getting history")
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com/feed/history", true)
	config.FileDump("HistoryRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, false)
	config.FileDump("HistoryRaw.json", jsonText, false)
	// Format into correct structure
	var jsonA HistJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("HistoryProcessed.json", string(text), false)
	
	contents := jsonA.Contents.TwoColumnBrowseResultsRenderer.Tabs[0]
	contentsB := contents.TabRenderer.Content.SectionListRenderer.Contents
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
		if videoJSON.Title.Runs != nil {
			
			// Views
			var views string
			var vidType string
			if videoJSON.ShortViewCountText.Runs == nil {
				 views = strings.Split(videoJSON.ShortViewCountText.SimpleText, " ")[0]
				 vidType = "Video"
			} else {
				views = videoJSON.ShortViewCountText.Runs[0].Text
				vidType = "Livestream"
			}
			
			// Published Time (history doesn't contain release date for some reason)
			var releaseDate string = "Unknown"
			
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
				ThumbnailLink: videoJSON.Thumbnail.Thumbnails[3].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				DirectLink: "",
				StartTime: videoJSON.NavigationEndpoint.WatchEndpoint.StartTimeSeconds,
				Type: youtube.VIDEO,
			}
			videos = append(videos, video)
			//fmt.Println(video.ThumbnailLink)
			go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, false)
		}
	}
	
	videoHolder := youtube.VideoHolder {
		Videos: videos,
		PageType: youtube.HISTORY,
		ContinuationToken: "",
		//ContinuationToken: contentsB[len(contentsB)-1].ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token,
	}
	//Print(contentsB[len(contentsB)-1].ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token)
	
	//fmt.Println("DONE Data")
	for i:=0; i<number; i++ {
		//fmt.Println("Doing thumbnails")
		_ = <- doneChan
	}
	return videoHolder
}

const PLAYLIST_URL string = "https://www.youtube.com/playlist?list="

func GetPlaylist(playlistId string, playlistName string) youtube.VideoHolder {
	config.LogEvent("Getting playlist " + playlistName)
	// Add to frecency file
	if playlistId != "WL" && playlistId != "LL" {
		config.LogEvent("Adding playlist to frecency file")
		AddToFile(playlistId, playlistName, youtube.HOME_DIR + youtube.CACHE_FOLDER + youtube.FRECENCY_PLAYLISTS_FILE)
	}
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML(PLAYLIST_URL + playlistId, true)
	config.FileDump("PlaylistRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, true)
	config.FileDump("PlaylistRaw.json", jsonText, false)
	// Format into correct structure
	var jsonA WLJSON
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("PlaylistProcessed.json", string(text), false)
	if jsonA.Contents.TwoColumnBrowseResultsRenderer.Tabs == nil {
		return youtube.VideoHolder{}
	}
	contents := jsonA.Contents.TwoColumnBrowseResultsRenderer.Tabs[0]
	contentsB := contents.TabRenderer.Content.SectionListRenderer.Contents
	contentsA := contentsB[0].ItemSectionRenderer.Contents[0].PlaylistVideoListRenderer.Contents
	videos := []youtube.Video{}
	
	var doneChan chan int = make(chan int)
	var err error
	_ = err
	var number int = 0
	for _, x := range contentsA {
		
		videoJSON := x.PlaylistVideoRenderer
		if videoJSON.Title.Runs != nil {
			
			// Views
			var views string = "Unknown"
			var vidType string = "Unknown"
			
			if videoJSON.VideoInfo.Runs != nil {
				if videoJSON.VideoInfo.Runs[1].Text == " watching" {
					views = videoJSON.VideoInfo.Runs[0].Text
					vidType = "Livestream"
				} else {
					views = videoJSON.VideoInfo.Runs[0].Text
					vidType = "Video"
				}
			}
			
			// Published Time
			var releaseDate string = "Livestream"
			if len(videoJSON.VideoInfo.Runs) > 2 {
				releaseDate = videoJSON.VideoInfo.Runs[2].Text
			}
			
			// Length
			var length string = "Livestream"
			if videoJSON.LengthText.SimpleText != "" {
				length = videoJSON.LengthText.SimpleText
			}
			
			// Remove params
			var playlistRemoveId string = ""
			var playlistRemoveParams string = ""
			for _, entry := range videoJSON.Menu.MenuRenderer.Items {
				if len(entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.ClientActions) > 0 {
					if entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.ClientActions[0].PlaylistRemoveVideosAction.SetVideoIds[0] != "" {
						playlistRemoveId = entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.ClientActions[0].PlaylistRemoveVideosAction.SetVideoIds[0]
						playlistRemoveParams =  entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.Params
					}
				}
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
				Channel: videoJSON.ShortBylineText.Runs[0].Text,
				ChannelID: videoJSON.ShortBylineText.Runs[0].NavigationEndpoint.CommandMetadata.WebCommandMetadata.URL,
				ThumbnailLink: videoJSON.Thumbnail.Thumbnails[3].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				DirectLink: "",
				StartTime: videoJSON.NavigationEndpoint.WatchEndpoint.StartTimeSeconds,
				PlaylistRemoveId: playlistRemoveId,
				PlaylistRemoveParams: playlistRemoveParams,
				Type: youtube.VIDEO,
				
			}
			videos = append(videos, video)
			go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, false)
		}
	}
	for i:=0; i<number; i++ {
		_ = <- doneChan
	}
	
	var pageType int
	if videos[0].PlaylistRemoveId == "" {
		pageType = youtube.OTHER_PLAYLIST
	} else {
		pageType = youtube.MY_PLAYLIST
	}
	
	videoHolder := youtube.VideoHolder {
		Videos: videos,
		PageType: pageType,
		PlaylistID: playlistId,
		ContinuationToken: contentsA[len(contentsA)-1].ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token,
	}
	
	return videoHolder
}

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
				Title: playlistJSON.Title.SimpleText,
				LastUpdated: lastUpdated,
				NumVideos: numVideos,
				Channel: author,
				Visibility: visibility,
				Id: playlistJSON.PlaylistID,
				ThumbnailLink: playlistJSON.Thumbnail.Thumbnails[0].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				Type: typeA,
			}
			playlists = append(playlists, playlist)
			go network.DownloadThumbnail(playlist.ThumbnailLink, playlist.ThumbnailFile, false, doneChan, false)
		}
	}
	for i:=0; i<number; i++ {
		_ = <- doneChan
	}
	
	holder := youtube.VideoHolder {
		Videos: playlists,
		PageType: youtube.LIBRARY,
		ContinuationToken: "",
	}
	
	return holder
}

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

func GetVideoPage(videoID string, playbackTrackingFilename string, skipThumbnails bool) (youtube.VideoPage, youtube.VideoHolder) {
	config.LogEvent("Getting video page " + videoID)
	// Get JSON text from the HTML
	var fullHTML string = network.GetHTML("https://www.youtube.com/watch?v=" + videoID, true)
	config.FileDump("VideoPageRaw.html", fullHTML, false)
	ytInitialData, initialPlayerResponse := network.ExtractJSONVideoPage(fullHTML)
	config.FileDump("VideoPageRawYTInitialData.json", ytInitialData, false)
	config.FileDump("VideoPageRawInitialPlayerResponse.json", initialPlayerResponse, false)
	// Format into correct structure
	var initialData VidPageInitialData
	var playerResponse VidPagePlayerResp
	
	if err := json.Unmarshal([]byte(ytInitialData), &initialData); err != nil {
		panic(err)
	}
	
	if err := json.Unmarshal([]byte(initialPlayerResponse), &playerResponse); err != nil {
		panic(err)
	}
	
	textInitialData, _ := json.MarshalIndent(initialData, "", "  ")
	textPlayerResponse, _ := json.MarshalIndent(playerResponse, "", "  ")
	config.FileDump("VideoPageProcessedYTInitialData.json", string(textInitialData), false)
	config.FileDump("VideoPageProcessedInitialPlayerResponse.json", string(textPlayerResponse), false)
	
	// First get the main video info
	primaryVideoInfo := initialData.Contents.TwoColumnWatchNextResults.Results.Results.Contents[0].VideoPrimaryInfoRenderer
	
	var subStatus string = "Unsubbed"
	var subParam string
	var unSubParam string
	if len(playerResponse.Annotations) > 0 {
		subInfo := playerResponse.Annotations[0].PlayerAnnotationsExpandedRenderer.FeaturedChannel.SubscribeButton.SubscribeButtonRenderer
		unSubInfo := subInfo.ServiceEndpoints[1].SignalServiceEndpoint.Actions[0].OpenPopupAction.Popup.ConfirmDialogRenderer.ConfirmButton.ButtonRenderer.ServiceEndpoint
		if subInfo.Subscribed {
			subStatus = "Subbed"
		}
		subParam = subInfo.ServiceEndpoints[0].SubscribeEndpoint.Params
		unSubParam = unSubInfo.UnsubscribeEndpoint.Params
	} else {
		
		if len(playerResponse.Endscreen.EndscreenRenderer.Elements) == 0 || len(playerResponse.Endscreen.EndscreenRenderer.Elements[0].EndscreenElementRenderer.HovercardButton.SubscribeButtonRenderer.ServiceEndpoints) == 0 {
			subInfo := playerResponse.PlayerConfig.WebPlayerConfig.WebPlayerActionsPorting
			subParam = subInfo.SubscribeCommand.SubscribeEndpoint.Params
			unSubParam = subInfo.UnsubscribeCommand.UnsubscribeEndpoint.Params
			// THIS CAN'T JUST BE 2, NEED TO ITERATE THROUGH AND CHECK ALL OF THEM
			if initialData.FrameworkUpdates.EntityBatchUpdate.Mutations[2].Payload.SubscriptionStateEntity.Subscribed {
				subStatus = "Subbed"
			}
			
		} else {
			subInfo := playerResponse.Endscreen.EndscreenRenderer.Elements[0].EndscreenElementRenderer.HovercardButton.SubscribeButtonRenderer
			if subInfo.Subscribed {
				subStatus = "Subbed"
			}
			subParam = subInfo.ServiceEndpoints[0].SubscribeEndpoint.Params
			unSubParam = subInfo.ServiceEndpoints[1].SignalServiceEndpoint.Actions[0].OpenPopupAction.Popup.ConfirmDialogRenderer.ConfirmButton.ButtonRenderer.ServiceEndpoint.UnsubscribeEndpoint.Params
		}
		
	}
	
	if subParam == "" {
		panic("Empty sub param")
	}
	if unSubParam == "" {
		panic("Empty unsub param")
	}
	
	addLikeInfo := primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.LikeButtonViewModel.LikeButtonViewModel.ToggleButtonViewModel.ToggleButtonViewModel.DefaultButtonViewModel.ButtonViewModel
	if addLikeInfo.IconName != "LIKE" || addLikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.CommandMetadata.WebCommandMetadata.ApiURL != "/youtubei/v1/like/like" {
		panic("Add like misplaced")
	}
	
	removeLikeInfo := primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.LikeButtonViewModel.LikeButtonViewModel.ToggleButtonViewModel.ToggleButtonViewModel.ToggledButtonViewModel.ButtonViewModel
	if removeLikeInfo.IconName != "LIKE" || removeLikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.CommandMetadata.WebCommandMetadata.ApiURL != "/youtubei/v1/like/removelike" {
		panic("Remove like misplaced")
	}
	
	addDislikeInfo := primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.DislikeButtonViewModel.DislikeButtonViewModel.ToggleButtonViewModel.ToggleButtonViewModel.DefaultButtonViewModel.ButtonViewModel
	if addDislikeInfo.IconName != "DISLIKE" || addDislikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.CommandMetadata.WebCommandMetadata.ApiURL != "/youtubei/v1/like/dislike" {
		panic("Add dislike misplaced")
	}
	
	removeDislikeInfo := primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.DislikeButtonViewModel.DislikeButtonViewModel.ToggleButtonViewModel.ToggleButtonViewModel.ToggledButtonViewModel.ButtonViewModel
	if removeDislikeInfo.IconName != "DISLIKE" || removeDislikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.CommandMetadata.WebCommandMetadata.ApiURL != "/youtubei/v1/like/removelike" {
		panic("Remove dislike misplaced")
	}
	
	mainVideo := youtube.VideoPage {
		Title: playerResponse.VideoDetails.Title,
		Views: primaryVideoInfo.ViewCount.VideoViewCountRenderer.ViewCount.SimpleText,
		ViewsShort: primaryVideoInfo.ViewCount.VideoViewCountRenderer.ShortViewCount.SimpleText,
		VidType: "",
		ReleaseDate: primaryVideoInfo.DateText.SimpleText,
		ReleaseDateShort: primaryVideoInfo.RelativeDateText.SimpleText,
		Length: playerResponse.VideoDetails.LengthSeconds,
		Likes: primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.LikeButtonViewModel.LikeButtonViewModel.ToggleButtonViewModel.ToggleButtonViewModel.DefaultButtonViewModel.ButtonViewModel.Title,
		Id: playerResponse.VideoDetails.VideoID,
		Channel: playerResponse.VideoDetails.Author,
		ChannelID: playerResponse.VideoDetails.ChannelID,
		ChannelThumbnailLink: initialData.Contents.TwoColumnWatchNextResults.Results.Results.Contents[1].VideoSecondaryInfoRenderer.Owner.VideoOwnerRenderer.Thumbnail.Thumbnails[2].URL,
		ChannelThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + "mainChannel.png",
		ThumbnailLink: playerResponse.VideoDetails.Thumbnail.Thumbnails[len(playerResponse.VideoDetails.Thumbnail.Thumbnails)-1].URL,
		ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + "main.png",
		DirectLink: "",
		Description: playerResponse.VideoDetails.ShortDescription,
		SubStatus: subStatus,
		SubParam: subParam,
		UnSubParam: unSubParam,
		LikeStatus: primaryVideoInfo.VideoActions.MenuRenderer.TopLevelButtons[0].SegmentedLikeDislikeButtonViewModel.LikeButtonViewModel.LikeButtonViewModel.LikeStatusEntity.LikeStatus,
		AddLikeParam: addLikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.LikeEndpoint.LikeParams,
		RemoveLikeParam: removeLikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.LikeEndpoint.RemoveLikeParams,
		AddDislikeParam: addDislikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.LikeEndpoint.DislikeParams,
		RemoveDislikeParam: removeDislikeInfo.OnTap.SerialCommand.Commands[1].InnertubeCommand.LikeEndpoint.RemoveLikeParams,
		VideoStatsPlaybackURL: playerResponse.PlaybackTracking.VideoStatsPlaybackURL.BaseURL,
		VideoStatsWatchtimeURL: playerResponse.PlaybackTracking.VideoStatsWatchtimeURL.BaseURL,
	}
	
	var doneChan chan int = make(chan int)
	
	// Download main video thumbnail and channel thumbnail
	if !skipThumbnails {
		go network.DownloadThumbnail(mainVideo.ThumbnailLink, mainVideo.ThumbnailFile, false, doneChan, true)
		go network.DownloadThumbnail(mainVideo.ChannelThumbnailLink, mainVideo.ChannelThumbnailFile, false, doneChan, true)
		_ = <- doneChan
		_ = <- doneChan
	}
	
	// Then get the suggestions
	videos := []youtube.Video{}
	
	contents := initialData.Contents.TwoColumnWatchNextResults.SecondaryResults.SecondaryResults.Results[1].ItemSectionRenderer.Contents
	
	var number int = 0
	
	for _, entry := range contents {
		if entry.CompactVideoRenderer.VideoID != "" {
		
			views := entry.CompactVideoRenderer.ShortViewCountText.SimpleText
			vidType := "Video"
			if len(entry.CompactVideoRenderer.ShortViewCountText.Runs) != 0 {
				views = entry.CompactVideoRenderer.ShortViewCountText.Runs[0].Text
				vidType = "Livestream"
			}
			
			length := "Unknown"
			if entry.CompactVideoRenderer.LengthText.SimpleText != "" {
				length = entry.CompactVideoRenderer.LengthText.SimpleText
			}
			
			video := youtube.Video {
				Title: entry.CompactVideoRenderer.Title.SimpleText,
				Views: views,
				VidType: vidType,
				ReleaseDate: entry.CompactVideoRenderer.PublishedTimeText.SimpleText,
				Length: length,
				Id: entry.CompactVideoRenderer.VideoID,
				Channel: entry.CompactVideoRenderer.ShortBylineText.Runs[0].Text,
				ChannelID: entry.CompactVideoRenderer.ShortBylineText.Runs[0].NavigationEndpoint.CommandMetadata.WebCommandMetadata.URL,
				ThumbnailLink: entry.CompactVideoRenderer.Thumbnail.Thumbnails[1].URL,
				ThumbnailFile: youtube.HOME_DIR + THUMBNAIL_DIR + strconv.Itoa(number) + ".png",
				DirectLink: "",
				StartTime: entry.CompactVideoRenderer.NavigationEndpoint.WatchEndpoint.StartTimeSeconds,
				Type: youtube.VIDEO,
			}
			number++
			videos = append(videos, video)
			
			if !skipThumbnails {
				go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, true)
			}
		}
	}
	
	videoHolder := youtube.VideoHolder {
		Videos: videos,
		PageType: youtube.VIDEO_PAGE,
		ContinuationToken: "",
	}
	
	//os.WriteFile("VideoPageProcessedYTInitialData.json", textInitialData, 0666)
	//os.WriteFile("VideoPageProcessedInitialPlayerResponse.json", textPlayerResponse, 0666)
	
	// Chapters
	var chaptersString string = ""
	if initialData.PlayerOverlays.PlayerOverlayRenderer.DecoratedPlayerBarRenderer.DecoratedPlayerBarRenderer.PlayerBar.MultiMarkersPlayerBarRenderer.MarkersMap != nil {
		chapters := initialData.PlayerOverlays.PlayerOverlayRenderer.DecoratedPlayerBarRenderer.DecoratedPlayerBarRenderer.PlayerBar.MultiMarkersPlayerBarRenderer.MarkersMap[0].Value.Chapters
		
		for _, chapter := range chapters {
			chaptersString = chaptersString + fmt.Sprintf("%sDELIM%s\n", chapter.ChapterRenderer.Title.SimpleText, strconv.Itoa(chapter.ChapterRenderer.TimeRangeStartMillis/1000))
		}
	}
	
	if skipThumbnails {
		//Print("saving to file")
		//os.WriteFile("THISraw.json", []byte(initialPlayerResponse), 0666)
		//thing, _ := json.MarshalIndent(initialData, "", "  ")
		//os.WriteFile("THISdone.json", []byte(string(thing)), 0666)
		//Print("saved to file")
		//var dirName string = "/tmp/" + strconv.Itoa(os.Getpid())
		//os.WriteFile(dirName + "/playbackTracking" + videoID + ".txt", []byte(mainVideo.VideoStatsPlaybackURL + "\n" + mainVideo.VideoStatsWatchtimeURL), 0666)
		//os.WriteFile(dirName + "/chapters" + videoID + ".txt", []byte(chaptersString), 0666)
		fmt.Println(mainVideo.VideoStatsPlaybackURL)
		fmt.Println(mainVideo.VideoStatsWatchtimeURL)
		fmt.Println(chaptersString)
	}
	
	if !skipThumbnails {
		for i:=0; i<number; i++ {
			_ = <- doneChan
		}
	}
	
	return mainVideo, videoHolder
}
