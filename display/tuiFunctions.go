package display

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"gotube/download"
	"gotube/config"
	"gotube/mpv"
	"gotube/youtube"
	"os"
	"os/exec"
	"strconv"
)

// This file contains meta functions related to the various on-keypress actions that can occur in the grid. Split into different categories as not all screens will have playlists for example

// Returns the currently selected video/playlist given the current selection
func getCurSelVid(content MainContent) youtube.Video {
	return content.GetVidHolder().Videos[content.getCurSel().Index]
}

// General functions not specific to either videos or playlists, and always avalible
func handleGeneralFunctions(key tcell.Key, r rune, mod tcell.ModMask, content MainContent) (int, []string) {
	// Exit
	if key == tcell.KeyEscape || key == tcell.KeyCtrlC || r == 'q' || r == 'Q' {
		return youtube.EXIT, []string{""}
	// Copy linx (share)
	} else if r == 's' {
		copyLink(content.getScreen(), getCurSelVid(content).Id, getCurSelVid(content).StartTime, getCurSelVid(content).Type)
	} else if mod == tcell.ModAlt && r == 'f' {
		openInBrowser(getCurSelVid(content).Id, getCurSelVid(content).Type)
	// Go to channel
	} else if r == 'c' {
		Print("Go to channel")
	// Download
	} else if r == 'd' {
		Print("Download")
	// Switch focus to search box
	} else if key == tcell.KeyTab {
		ret, data := FocusSearchBox(content, false)
		if ret != youtube.NONE {
			return ret, data
		}
	} else if r == '/' {
		currentSearchTerm = "/"
		cursorLoc = 1
		ret, data := FocusSearchBox(content, false)
		if ret != youtube.NONE {
			return ret, data
		}
	}
	return youtube.NONE, nil
}

// Functions that only apply to videos
func handleVideoFunctions(key tcell.Key, r rune, mod tcell.ModMask, content MainContent) (int, []string) {
	//Print("Handling video functions")
	// Launch video in foreground
	if key == tcell.KeyEnter && mod == tcell.ModAlt && getCurSelVid(content).Type == youtube.VIDEO {
		var timestamp string = playVideoForeground(content, false)
		return youtube.VIDEO_PAGE, []string{getCurSelVid(content).Id, timestamp}
	// Launch video in the background
	} else if key == tcell.KeyEnter && mod == tcell.ModNone && getCurSelVid(content).Type == youtube.VIDEO {
		playVideoBackground(content, false)
	// Launch video with quality options in foreground
	} else if r == '#' && mod == tcell.ModAlt && getCurSelVid(content).Type == youtube.VIDEO {
		var timestamp string = playVideoForeground(content, true)
		return youtube.VIDEO_PAGE, []string{getCurSelVid(content).Id, timestamp}
	// Launch video with quality options in background
	} else if r == '#' && mod == tcell.ModNone && getCurSelVid(content).Type == youtube.VIDEO {
		playVideoBackground(content, true)
	// Add to Watch Later
	} else if r == 'w' && getCurSelVid(content).Type == youtube.VIDEO {
		addToPlaylist(content.getScreen(), getCurSelVid(content).Id, "WL", "Watch later")
	// Add to playlist
	} else if r == 'a' && getCurSelVid(content).Type == youtube.VIDEO {
		addToPlaylistOptions(content)
	// Remove from playlist
	} else if r == 'r' && content.GetVidHolder().PageType == youtube.MY_PLAYLIST {
		removeFromPlaylist(content)
	}

	return youtube.NONE, nil
}

// Functions that only apply to playlists
func handlePlaylistFunctions(key tcell.Key, r rune, mod tcell.ModMask, content MainContent) (int, []string) {
	// Open playlist
	if key == tcell.KeyEnter && mod == tcell.ModNone && (getCurSelVid(content).Type == youtube.OTHER_PLAYLIST || getCurSelVid(content).Type == youtube.MY_PLAYLIST) {
		return youtube.GET_PLAYLIST, []string{getCurSelVid(content).Id, getCurSelVid(content).Title}
	// Open playlist
	} else if key == tcell.KeyEnter && mod == tcell.ModAlt && (getCurSelVid(content).Type == youtube.OTHER_PLAYLIST || getCurSelVid(content).Type == youtube.MY_PLAYLIST) {
		cmd := exec.Command("nohup", config.ActiveConfig.Term, "-e", "gotube", "-p", getCurSelVid(content).Id, getCurSelVid(content).Title)
		cmd.Start()
	// Save to library
	} else if r == 'a' && (getCurSelVid(content).Type == youtube.OTHER_PLAYLIST || getCurSelVid(content).Type == youtube.MY_PLAYLIST) {
		addToLibrary(content.getScreen(), getCurSelVid(content).Id, getCurSelVid(content).Title)
	// Remove from library
	} else if r == 'r' && content.GetVidHolder().PageType == youtube.LIBRARY && getCurSelVid(content).Type == youtube.OTHER_PLAYLIST {
		removeFromLibrary(content)
	}
	return youtube.NONE, nil
}

// Below are the individual functions which handle keypress requests

func playVideoBackground(content MainContent, qualitySelection bool) {
	timestamp := getTimestampFilename()
	StartLoading(content.getScreen())
	playVideo(content, qualitySelection, timestamp)
	EndLoading()
}

func playVideoForeground(content MainContent, qualitySelection bool) string {
	timestamp := getTimestampFilename()
	StartLoading(content.getScreen()) // End Loading called in main, it's not missing
	content.setCurSel(playVideo(content, qualitySelection, timestamp))
	return timestamp
}

func addToLibrary(screen tcell.Screen, playlistId string, playlistName string) {
	StartLoading(screen)
	ok := download.AddToLibrary(playlistId)
	EndLoading()
	if ok {
		drawStatusBar(screen, []string{"Added playlist " + playlistName + " to library"})
	} else {
		drawStatusBar(screen, []string{"Error, could not add to library"})
	}
	screen.Sync()
}

func removeFromLibrary(content MainContent) {
	screen := content.getScreen()
	StartLoading(screen)
	ok := download.RemoveFromLibrary(getCurSelVid(content).Id)
	EndLoading()
	if ok {
		var curIndex int = content.getCurSel().Index
		editedVideos := append(content.GetVidHolder().Videos[:curIndex], content.GetVidHolder().Videos[curIndex+1:]...)
		content.SetVideosList(editedVideos)
		content.calcSizing()
		content.recalibrate()
		content.redraw(REDRAW_IMAGES, HIDE_CURSOR)
	} else {
		drawStatusBar(screen, []string{"Error, could not remove from library"})
	}
	screen.Sync()
}

func addToPlaylistOptions(content MainContent) {
	var videoId string = getCurSelVid(content).Id

	playlistOptionsMap, playlistOptionsList := download.GetAddToPlaylist(videoId)
	chosen := selectionTUI(content, playlistOptionsList, false)
	if chosen == "" {
		return
	}
	addToPlaylist(content.getScreen(), videoId, playlistOptionsMap[chosen], chosen)
}

func addToPlaylist(screen tcell.Screen, videoId, playlistId, playlistName string) {
	StartLoading(screen)
	ok := download.AddToPlaylist(videoId, playlistId)
	EndLoading()
	if ok {
		drawStatusBar(screen, []string{"Added to " + playlistName})
	} else {
		drawStatusBar(screen, []string{"Error, could not add to playlist"})
	}
	screen.Sync()
}

func removeFromPlaylist(content MainContent) {
	screen := content.getScreen()
	StartLoading(screen)
	ok := download.RemoveFromPlaylist(getCurSelVid(content).Id, content.GetVidHolder().PlaylistID, getCurSelVid(content).PlaylistRemoveId, getCurSelVid(content).PlaylistRemoveParams)
	EndLoading()

	if ok {
		var curIndex int = content.getCurSel().Index
		editedVideos := append(content.GetVidHolder().Videos[:curIndex], content.GetVidHolder().Videos[curIndex+1:]...)
		content.SetVideosList(editedVideos)
		content.calcSizing()
		content.recalibrate()
		content.redraw(REDRAW_IMAGES, HIDE_CURSOR)
	} else {
		drawStatusBar(screen, []string{"Error, could not remove from playlist"})
	}
	screen.Sync()
}

func copyLink(screen tcell.Screen, id string, startTime int, itemType int) {
	if itemType == youtube.VIDEO {
		copyToClipboard("https://www.youtube.com/watch?v=" + id + "&t=" + strconv.Itoa(startTime))
	} else if itemType == youtube.MY_PLAYLIST || itemType == youtube.OTHER_PLAYLIST {
		copyToClipboard("https://www.youtube.com/playlist?list=" + id)
	}
	drawStatusBar(screen, []string{"Copied link to clipboard"})
	screen.Sync()
}

func getExtension(screen tcell.Screen, videosHolder youtube.VideoHolder) youtube.VideoHolder {
	StartLoading(screen)
	videosHolder = download.GetPlaylistContinuation(videosHolder, videosHolder.ContinuationToken)
	EndLoading()
	return videosHolder
}

func openInBrowser(id string, contentType int) {
	var link string
	if contentType == youtube.VIDEO {
		link = "https://www.youtube.com/watch?v=" + id
	} else if contentType == youtube.MY_PLAYLIST || contentType == youtube.OTHER_PLAYLIST {
		link = "https://www.youtube.com/playlist?list=" + id
	}
	cmd := exec.Command("nohup", config.ActiveConfig.Browser, link)
	cmd.Start()
}

//func DetachVideo(title string, channel string, startTime string, startNum string, folderName string, quality string)

func playVideo(content MainContent, qualitySelection bool, timestamp string) CurSelection {
	//var qualityOptions map[string]youtube.Format = download.GetDirectLinks(getCurSelVid(content).Id)
	qualityOptions := []string{"360p", "720p", "1080p", "1440p", "2160p"}
	mpv.WritePlaylistFile(content.GetVidHolder())

	var desiredQuality string = "720p"
	var curSel CurSelection
	if qualitySelection {
		desiredQuality = selectionTUI(content, qualityOptions, true)
		if desiredQuality == "" {
			return curSel
		}
	}
	video := getCurSelVid(content)

	var windowWidth, windowHeight, windowPosX, windowPosY int = getWindowSizeAndPosition()
	var geometryArgument string = fmt.Sprintf("%dx%d+%d+%d", windowWidth, windowHeight, windowPosX, windowPosY)

	go mpv.DetachVideo(video.Title, video.Channel, strconv.Itoa(video.StartTime), strconv.Itoa(content.getCurSel().Index), "/tmp/" + strconv.Itoa(os.Getpid()), desiredQuality, geometryArgument)
	return curSel
}
