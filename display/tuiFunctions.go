package display

import (
	"github.com/gdamore/tcell/v2"
	"gotube/youtube"
	"strconv"
	"gotube/download"
	"os"
	"gotube/mpv"
)

// This file contains meta functions related to the various on-keypress actions that can occur in the grid. Split into different categories as not all screens will have playlists for example

// Returns the currently selected video/playlist given the current selection
func getCurSelVid(content MainContent) youtube.Video {
	return content.GetVidHolder().Videos[content.getCurSel().Index]
}

// General functions not specific to either videos or playlists, and always avalible
func handleGeneralFunctions(key tcell.Key, r rune, content MainContent) (int, []string) {
	// Exit
	 if key == tcell.KeyEscape || key == tcell.KeyCtrlC || r == 'q' || r == 'Q' {
		return youtube.EXIT, []string{""}
	// Copy linx (share)
	} else if r == 's' {
		copyLink(content.getScreen(), getCurSelVid(content).Id, getCurSelVid(content).Type)
	// Go to channel
	} else if r == 'c' {
		Print("Go to channel")
	// Download
	} else if r == 'd' {
		Print("Download")
	// Switch focus to search box
	} else if key == tcell.KeyTab || r == '/' {
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
	} else if key == tcell.KeyEnter && mod == tcell.ModNone && getCurSelVid(content).Type == youtube.VIDEO  {
		playVideoBackground(content, false)
	// Launch video with quality options in foreground
	} else if r == '#' && mod == tcell.ModAlt && getCurSelVid(content).Type == youtube.VIDEO  {
		var timestamp string = playVideoForeground(content, true)
		return youtube.VIDEO_PAGE, []string{getCurSelVid(content).Id, timestamp}
	// Launch video with quality options in background
	} else if r == '#' && mod == tcell.ModNone && getCurSelVid(content).Type == youtube.VIDEO  {
		playVideoBackground(content, true)
	// Add to Watch Later
	} else if r == 'w' && getCurSelVid(content).Type == youtube.VIDEO  {
		addToPlaylist(content.getScreen(), getCurSelVid(content).Id, "WL", "Watch later")
	// Add to playlist
	} else if r == 'a' && getCurSelVid(content).Type == youtube.VIDEO  {
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
	
	var playlistOptions map[string]string = download.GetAddToPlaylist(videoId)
	chosen := selectionTUI(content, sliceFromMap[string](playlistOptions))
	addToPlaylist(content.getScreen(), videoId, playlistOptions[chosen], chosen)
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

func copyLink(screen tcell.Screen, id string, itemType int) {
	if itemType == youtube.VIDEO {
		copyToClipboard("https://www.youtube.com/watch?v=" + id)
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

//func DetachVideo(title string, channel string, startTime string, startNum string, folderName string, quality string)

func playVideo(content MainContent, qualitySelection bool, timestamp string) CurSelection {
	var qualityOptions map[string]youtube.Format = download.GetDirectLinks(getCurSelVid(content).Id)
	
	mpv.WritePlaylistFile(content.GetVidHolder())
	
	var desiredQuality string = "720p"
	var curSel CurSelection
	if qualitySelection {
		desiredQuality = selectionTUI(content, sliceFromMap[youtube.Format](qualityOptions))
	}
	video := getCurSelVid(content)
	go mpv.DetachVideo(video.Title, video.Channel, strconv.Itoa(video.StartTime), strconv.Itoa(content.getCurSel().Index), "/tmp/" + strconv.Itoa(os.Getpid()), desiredQuality)
	return curSel
}

