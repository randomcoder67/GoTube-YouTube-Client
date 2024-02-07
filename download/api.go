package download

import (
	"gotube/youtube"
	"gotube/config"
	"encoding/json"
	"strconv"
	"strings"
	"fmt"
	"math/rand"
	"gotube/download/network"
)

// This file contains various functions for interacting with the YouTube API

const API_KEY string = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
const PLAYLIST_ADD_URL = "https://www.youtube.com/youtubei/v1/browse/edit_playlist?key=" + API_KEY
const BROWSE_URL = "https://www.youtube.com/youtubei/v1/browse?key=" + API_KEY
const GET_ADD_TO_PLAYLIST_URL string = "https://www.youtube.com/youtubei/v1/playlist/get_add_to_playlist?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
const ADD_LIKE_URL string = "https://www.youtube.com/youtubei/v1/like/like?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
const REMOVE_LIKE_URL string = "https://www.youtube.com/youtubei/v1/like/removelike?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

type markWatchedKeys struct {
	cpn string
	docid string
	ns string
	el string
	uga string
	ver string
	st string
	cl string
	ei string
	plid string
	length string
	of string
	vm string
	cmt string
	et string
}

const cpnOptions string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

// Helper function for the process of marking a video as watched
func getCPN() string {
	var toReturn string = ""
	for i:=0; i<16; i++ {
		toReturn += string(cpnOptions[rand.Intn(63)])
	}
	return toReturn
}

// Extract info from the url contained in the page data
func extractInfo(url string, time string) markWatchedKeys {
	split := strings.Split(url, "&")
	var toReturn markWatchedKeys = markWatchedKeys{
		cpn: getCPN(),
		docid: strings.Split(split[1], "=")[1],
		ns: strings.Split(split[4], "=")[1],
		el: strings.Split(split[6], "=")[1],
		uga: strings.Split(split[len(split)-2], "=")[1],
		ver: "2",
		st: "0",
		cl: strings.Split(split[0], "=")[1],
		ei: strings.Split(split[2], "=")[1],
		plid: strings.Split(split[5], "=")[1],
		length: strings.Split(split[7], "=")[1],
		of: strings.Split(split[8], "=")[1],
		vm: strings.Split(split[len(split)-1], "=")[1],
		cmt: time + ".0",
		et: time + ".0",
	}
	
	return toReturn
}

// Mark a video as watched
func MarkWatched(videoId string, videoStatsPlaybackURL string, videoStatsWatchtimeURL string, time string) {
	config.LogEvent("Adding video: " + videoId + " to history with time: " + time)
	time = strings.Split(time, ".")[0]
	
	playbackData := extractInfo(videoStatsPlaybackURL, time)
	watchtimeData := extractInfo(videoStatsWatchtimeURL, time)
	//Print(watchtimeData.cmt)
	
	var playbackURL string = fmt.Sprintf("https://s.youtube.com/api/stats/playback?cl=%s&docid=%s&ei=%s&ns=%s&plid=%s&el=%s&len=%s&of=%s&uga=%s&vm=%s&ver=%s&cpn=%s&cmt=%s", playbackData.cl, playbackData.docid, playbackData.ei, playbackData.ns, playbackData.plid, playbackData.el, playbackData.length, playbackData.of, playbackData.uga, playbackData.vm, playbackData.ver, playbackData.cpn, playbackData.cmt)
	
	var watchtimeURL string = fmt.Sprintf("https://s.youtube.com/api/stats/watchtime?cl=%s&docid=%s&ei=%s&ns=%s&plid=%s&el=%s&len=%s&of=%s&uga=%s&vm=%s&ver=%s&cpn=%s&cmt=%s&st=%s&et=%s", watchtimeData.cl, watchtimeData.docid, watchtimeData.ei, watchtimeData.ns, watchtimeData.plid, watchtimeData.el, watchtimeData.length, watchtimeData.of, watchtimeData.uga, watchtimeData.vm, watchtimeData.ver, watchtimeData.cpn, watchtimeData.cmt, watchtimeData.st, watchtimeData.et)
	
	config.FileDump("PlaybackURLFinal.txt", playbackURL, false)
	config.FileDump("WatchtimeURLFinal.txt", playbackURL, false)
	
	network.GetHTML(playbackURL, true)
	network.GetHTML(watchtimeURL, true)
}

func getPLAddRemove(playlistId string) PLAddRemove {
	var fullHTML string = network.GetHTML(PLAYLIST_URL + playlistId, true)
	config.FileDump("PLAddRemoveRaw.html", fullHTML, false)
	var jsonText string = network.ExtractJSON(fullHTML, true)
	config.FileDump("PLAddRemoveRaw.json", jsonText, false)
	
	//os.WriteFile("pl.json", []byte(jsonText), 0666)
	
	var jsonA PLAddRemove
	if err := json.Unmarshal([]byte(jsonText), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("PLAddRemoveProcessed.json", string(text), false)
	
	return jsonA
}

// Add a playlist to library
func AddToLibrary(playlistId string) bool {
	config.LogEvent("Adding playlist: " + playlistId + " to library")
	jsonA := getPLAddRemove(playlistId)
	
	 jsonString := `{
		"context": {
			"client": {
				"clientName":"WEB",
				"clientVersion":"2.20231121.08.00"
			}
		},
		"target": {
			"playlistId":"PLAYLISTID"
		},
		"params":"PARAMS"
	}`
	
	jsonString = strings.ReplaceAll(strings.ReplaceAll(jsonString, "PLAYLISTID", playlistId), "PARAMS", jsonA.Header.PlaylistHeaderRenderer.SaveButton.ToggleButtonRenderer.DefaultServiceEndpoint.LikeEndpoint.LikeParams)
	
	status, response := network.PostRequestAPI(jsonString, ADD_LIKE_URL, "https://www.youtube.com/playlist?list=" + playlistId)
	
	config.FileDump("AddToLibraryResponse.json", response, false)
	
	if status == 200 {
		return true
	}
	return false
}

// Remove a playlist from library
func RemoveFromLibrary(playlistId string) bool {
	config.LogEvent("Removing playlist: " + playlistId + " from library")
	jsonA := getPLAddRemove(playlistId)
	
	 jsonString := `{
		"context": {
			"client": {
				"clientName":"WEB",
				"clientVersion":"2.20231121.08.00"
			}
		},
		"target": {
			"playlistId":"PLAYLISTID"
		},
		"params":"PARAMS"
	}`
	
	jsonString = strings.ReplaceAll(strings.ReplaceAll(jsonString, "PLAYLISTID", playlistId), "PARAMS", jsonA.Header.PlaylistHeaderRenderer.SaveButton.ToggleButtonRenderer.ToggledServiceEndpoint.LikeEndpoint.DislikeParams)
	
	status, response := network.PostRequestAPI(jsonString, REMOVE_LIKE_URL, "https://www.youtube.com/playlist?list=" + playlistId)
	
	config.FileDump("RemoveFromLibraryResponse.json", response, false)
	
	if status == 200 {
		return true
	}
	return false
}

func AddToPlaylist(videoID string, playlistID string) bool {
	config.LogEvent("Adding video: " + videoID + " to playlist: " + playlistID)
	jsonString := `{
		"context": {
			"client": {
				"clientName":"WEB",
				"clientVersion":"2.20231121.08.00"
			}
		},
		"actions": [
		{
			"addedVideoId":"VIDEOID",
			"action":"ACTION_ADD_VIDEO"
		}
		],
		"playlistId":"PLAYLISTID"
	}`
	
	jsonString = strings.ReplaceAll(strings.ReplaceAll(jsonString, "VIDEOID", videoID), "PLAYLISTID", playlistID)
	status, response := network.PostRequestAPI(jsonString, PLAYLIST_ADD_URL, "https://www.youtube.com/watch?v=" + videoID)
	
	config.FileDump("AddToPlaylistResponse.json", response, false)
	
	if status == 200 {
		return true
	}
	return false
}

func RemoveFromPlaylist(videoID string, playlistID string, removeID string, removeParams string) bool {
	config.LogEvent("Removing video: " + videoID + " from playlist: " + playlistID)
	jsonString := `{
		"context": {
			"client": {
				"clientName":"WEB",
				"clientVersion":"2.20231121.08.00"
			}
		},
		"actions": [
		{
			"action":"ACTION_REMOVE_VIDEO",
			"setVideoId":"REMOVEID"
		}
		],
		"params": "REMOVEPARAMS",
		"playlistId":"PLAYLISTID"
	}`
	
	jsonString = strings.ReplaceAll(strings.ReplaceAll(jsonString, "VIDEOID", videoID), "PLAYLISTID", playlistID)
	jsonString = strings.ReplaceAll(strings.ReplaceAll(jsonString, "REMOVEID", removeID), "REMOVEPARAMS", removeParams)
	status, response := network.PostRequestAPI(jsonString, PLAYLIST_ADD_URL, "https://www.youtube.com/playlist?list=" + playlistID)
	
	config.FileDump("RemoveFromPlaylistResponse.json", response, false)
	
	if status == 200 {
		return true
	}
	return false
}

func GetAddToPlaylist(videoID string) map[string]string {
	config.LogEvent("Getting AddToPlaylist information for video: " + videoID)
	jsonString := `{
		"context": {
			"client": {
				"clientName": "WEB",
				"clientVersion": "2.20231214.06.00"
			}
		},
		"videoIds": [
			"VIDEOID"
		],
		"excludeWatchLater": false
	}`
	
	jsonString = strings.ReplaceAll(jsonString, "VIDEOID", videoID)
	
	status, returnedJSONString := network.PostRequestAPI(jsonString, GET_ADD_TO_PLAYLIST_URL, "https://www.youtube.com/watch?v=" + videoID)
	config.FileDump("GetAddToPlaylistRaw.json", returnedJSONString, false)
	
	var infoJSON PlaylistAddDataJSON
	if err := json.Unmarshal([]byte(returnedJSONString), &infoJSON); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(infoJSON, "", "  ")
	config.FileDump("GetAddToPlaylistProcessed.json", string(text), false)
	
	_ = status
	
	// This does come in the correct order (i.e. most recently updated first), however it's then unsorted as I'm using a map - FIX THIS
	playlistsMap := make(map[string]string)
	contents := infoJSON.Contents[0].AddToPlaylistRenderer.Playlists
	for _, entry := range contents {
		info := entry.PlaylistAddToOptionRenderer
		playlistsMap[info.Title.SimpleText] = info.PlaylistID
	}
	return playlistsMap
}

func GetPlaylistContinuation(videosHolder youtube.VideoHolder, continuationToken string) youtube.VideoHolder {
	config.LogEvent("Getting playlist continuation for playlist: " + videosHolder.PlaylistID)
	videos := videosHolder.Videos
	
	jsonString := `{
	  "context": {
		"client": {
		  "clientName": "WEB",
		  "clientVersion": "2.20231214.06.00"
		},
		"user": {
		  "lockedSafetyMode": "false"
		},
		"request": {
		  "useSsl": "true"
		}
	  },
	  "continuation": "CONTINUE"
	}`
	
	jsonString = strings.ReplaceAll(jsonString, "CONTINUE", continuationToken)
	status, returnedJSONString := network.PostRequestAPI(jsonString, BROWSE_URL, "https://www.youtube.com/playlist?list=" + videosHolder.PlaylistID)
	
	config.FileDump("PlaylistContinuationRaw.json", returnedJSONString, false)
	
	_ = videos
	_ = status
	// Format into correct structure
	var jsonA ContinuationJSON
	if err := json.Unmarshal([]byte(returnedJSONString), &jsonA); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonA, "", "  ")
	config.FileDump("PlaylistContinuationProcessed.json", string(text), false)
	
	contents := jsonA.OnResponseReceivedActions[0].AppendContinuationItemsAction.ContinuationItems
	
	videosHolder.ContinuationToken = ""
	
	
	var doneChan chan int = make(chan int)
	var err error
	_ = err
	var oldNumber int = len(videos)
	var number int = len(videos)
	for _, x := range contents {
		
		videoJSON := x.PlaylistVideoRenderer
		continuationJSON := x.ContinuationItemRenderer
		
		if videoJSON.Title.Runs != nil {
			
			// Views
			var views string
			var vidType string
			if videoJSON.VideoInfo.Runs[1].Text == " watching" {
				 views = videoJSON.VideoInfo.Runs[0].Text
				vidType = "Livestream"
			} else {
				 views = videoJSON.VideoInfo.Runs[0].Text
				 vidType = "Video"
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
					if len(entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.ClientActions[0].PlaylistRemoveVideosAction.SetVideoIds) > 0 {
						playlistRemoveId = entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.ClientActions[0].PlaylistRemoveVideosAction.SetVideoIds[0]
						playlistRemoveParams =  entry.MenuServiceItemRenderer.ServiceEndpoint.PlaylistEditEndpoint.Params
					}
				}
			}
			
			if playlistRemoveId == "" {
				//Print("ERROR, no reomve ID")
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
				
			}
			videos = append(videos, video)
			go network.DownloadThumbnail(video.ThumbnailLink, video.ThumbnailFile, false, doneChan, false)
		} else if continuationJSON.ContinuationEndpoint.ContinuationCommand.Token != "" {
			videosHolder.ContinuationToken = continuationJSON.ContinuationEndpoint.ContinuationCommand.Token
		}
	}
	//fmt.Println("DONE Data")
	for i:=0; i<number-oldNumber; i++ {
		//fmt.Println("Doing thumbnails")
		_ = <- doneChan
	}

	videosHolder.Videos = videos
	return videosHolder
}

func GetDirectLinks(videoID string) map[string]youtube.Format {
	config.LogEvent("Getting direct links for video: " + videoID)
	structJSON := &network.PostJSON{
		VideoID: videoID,
		Params: "CgIQBg==",
		ContentCheckOK: true,
		RacyCheckOK: true,
	}
	
	structJSON.Context.Client.ClientName = "ANDROID"
	structJSON.Context.Client.ClientVersion = "17.31.35"
	structJSON.Context.Client.UserAgent = "com.google.android.youtube/17.31.35 (Linux; U; Android 11) gzip"
	structJSON.Context.Client.AndroidSDKVersion = 30 
	structJSON.Context.Client.HL = "en"
	structJSON.Context.Client.TimeZone = "UTC"
	structJSON.Context.Client.UTCOffsetMinutes = 0
	structJSON.PlaybackContext.ContentPlaybackContext.HTML5Preference = "HTML5_PREF_WANTS"
	
	var jsonString string = network.PostRequest(structJSON)
	config.FileDump("GetDirectLinksRaw.json", jsonString, false)
	
	var jsonFormats DLResponse
	if err := json.Unmarshal([]byte(jsonString), &jsonFormats); err != nil {
		panic(err)
	}
	
	text, _ := json.MarshalIndent(jsonFormats, "", " ")
	config.FileDump("GetDirectLinksProcessed.json", string(text), false)
	
	qualityOptions := make(map[string]youtube.Format)
	
	for _, entry := range jsonFormats.StreamingData.Formats {
		if entry.Itag == 22 {
			qualityOptions["720p"] = youtube.Format{VideoURL: entry.URL, AudioURL: "combined"}
		} else if entry.Itag == 18 {
			qualityOptions["360p"] = youtube.Format{VideoURL: entry.URL, AudioURL: "combined"}
		}
	}
	
	var audioLinkM4A string
	var audioLinkOpus string
	for _, entry := range jsonFormats.StreamingData.AdaptiveFormats {
		if entry.Itag == 140 {
			audioLinkM4A = entry.URL
		} else if entry.Itag == 251 {
			audioLinkOpus = entry.URL
		}
	}
	
	for _, entry := range jsonFormats.StreamingData.AdaptiveFormats {
		if entry.Itag == 313 {
			qualityOptions["2160p"] = youtube.Format{VideoURL: entry.URL, AudioURL: audioLinkOpus}
		} else if entry.Itag == 271 {
			qualityOptions["1440p"] = youtube.Format{VideoURL: entry.URL, AudioURL: audioLinkOpus}
		} else if entry.Itag == 248 || entry.Itag == 303 {
			qualityOptions["1080p"] = youtube.Format{VideoURL: entry.URL, AudioURL: audioLinkOpus}
		} else if entry.Itag == 397 {
			qualityOptions["480p"] = youtube.Format{VideoURL: entry.URL, AudioURL: audioLinkM4A}
		}
	}
	
	//Print("Length: " + strconv.Itoa(len(qualityOptions)))
	// If it's a livestream, just return that
	if jsonFormats.StreamingData.HLSManifestURL != "" {
		qualityOptions = make(map[string]youtube.Format)
		qualityOptions["1080p"] = youtube.Format{VideoURL: jsonFormats.StreamingData.HLSManifestURL, AudioURL: "combined"}
		//Print("Livestream: " + jsonFormats.StreamingData.HLSManifestURL)
	}
			
	return qualityOptions
}
