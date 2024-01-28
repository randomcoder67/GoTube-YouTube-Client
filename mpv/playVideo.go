package mpv

import (
	"fmt"
	"os/exec"
	"bytes"
	"gotube/youtube"
	"gotube/download"
	"os"
)

var _ = fmt.Println
var _ = exec.Command
var _ youtube.Video
var _ = os.WriteFile

/*
func parseMPVVidTime(input string) string {
	splitA := strings.Split(input, "AV:")
	split := strings.Split(splitA[len(splitA)-1], ":")
	var hours string = split[0][1:]
	var minutes string = split[1]
	var seconds string = split[2][:2]
	fmt.Printf("H:%s,M:%s,S:%s\n", hours, minutes, seconds)
	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		panic(err)
	}
	minutesInt, err := strconv.Atoi(minutes)
	if err != nil {
		panic(err)
	}
	secondsInt, err := strconv.Atoi(seconds)
	if err != nil {
		panic(err)
	}
	
	return strconv.Itoa(hoursInt*3600 + minutesInt*60 + secondsInt)
}
*/

func GetVideoData(videoId string) {
	//Print("IN GET VIDEO DATA: " + videoId)
	download.GetVideoPage(videoId, "/dev/null", true)
	//Print("DONE GET VIDEO DATA")
}

func GetQualityLinks(videoId string, quality string) () {
	qualityOptions := download.GetDirectLinks(videoId)
	var videoLink, audioLink string
	result, ok := qualityOptions[quality]
	if !ok {
		result = qualityOptions["720p"]
		videoLink = result.VideoURL
		audioLink = result.AudioURL
	} else {
		videoLink = result.VideoURL
		audioLink = result.AudioURL
	}
	
	fmt.Printf("%s\n%s\n", videoLink, audioLink)
}

func MarkWatched(videoId string, finalTime string, videoStatsPlaybackURL string, videoStatsWatchtimeURL string) () {
	/*
	const MAX_TRIES int = 3
	var i int = 0
	var videoStatsPlaybackURL, videoStatsWatchtimeURL string
	for {
		if _, err := os.Stat(fileName); err == nil {
			i++
			dat, err := os.ReadFile(fileName)
			if err != nil {
				if i > MAX_TRIES {
					Print("PANIC")
					panic(err)
				}
				continue
			}
			split := strings.Split(string(dat), "\n")
			videoStatsPlaybackURL = split[0]
			videoStatsWatchtimeURL = split[1]
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	*/
	//Print("IN MARK WATCHED")
	//Print(videoId)
	//Print(finalTime)
	//Print(videoStatsPlaybackURL)
	//Print(videoStatsWatchtimeURL)
	download.MarkWatched(videoId, videoStatsPlaybackURL, videoStatsWatchtimeURL, finalTime)
}

func StartPlayback(title string, channel string, startTime string, startNum string, folderName string, quality string) {
	mpvCommandArgs := []string{"--title=" + title + " - " + channel, "--start=" + startTime, "--playlist-start=" + startNum, "--script=" + youtube.HOME_DIR + "/.local/bin/gotube.lua", "--script-opts=gotube-folderName=" + folderName + ",gotube-quality=" + quality, "--playlist=" + folderName + "/playlist.m3u"}
	
	var thing string = ""
	for _, x := range mpvCommandArgs {
		thing += x
	}
	//os.WriteFile("mpv.command", []byte(thing), 0666)
	
	
	mpvVideoCommand := exec.Command("mpv", mpvCommandArgs...)
	//os.WriteFile("mpv.command2", []byte(mpvVideoCommand.String()), 0666)
	
	var outb, errb bytes.Buffer
	mpvVideoCommand.Stdout = &outb
	mpvVideoCommand.Stderr = &errb
	
	err := mpvVideoCommand.Run()
	if err != nil {
		Print(err.Error())
		//panic(err)
	}
	
	//os.WriteFile("mpv.out", []byte(outb.String()), 0666)
	//os.WriteFile("mpv.err", []byte(errb.String()), 0666)
	//if exitErr, ok := err.(*exec.ExitError); ok {
	//	exitCode := exitErr.ExitCode()
	//	Print("Exit Code: " + strconv.Itoa(exitCode))
	//}
}
