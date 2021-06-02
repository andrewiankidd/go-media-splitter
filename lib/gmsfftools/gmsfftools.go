package gmsfftools

import (
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/andrewiankidd/go-media-splitter/lib/gmstypes"
	"github.com/andrewiankidd/go-media-splitter/lib/gmsutils"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var (
	Verbose *bool
)

type FFFileInfo = gmstypes.FFFileInfo
type FFFrameset = gmstypes.FFFrameset

func SortFrames(ss []FFFrameset) (ret []FFFrameset) {

	// sort input array by duration
	// DESC (longest first)
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Duration > ss[j].Duration
	})

	// return sorted results
	return ss
}

func ReduceFramesets(ss []string, minDuration float64) (ret []FFFrameset) {

	// iterate each entry of the input array
	for i := range ss {

		// we want only every even entry,
		// so we can reduce into pairs
		if i%2 == 0 {

			// parse start string val into float
			st, stErr := strconv.ParseFloat(ss[i], 64)
			if stErr != nil {
				log.Fatal(stErr)
			}

			// parse end string val into float
			en, enErr := strconv.ParseFloat(ss[i+1], 64)
			if enErr != nil {
				log.Fatal(enErr)
			}

			// build as new framedata
			var b = FFFrameset{Start: st, End: en, Duration: (en - st)}

			if minDuration == 0 || b.Duration > minDuration {
				// append to return object
				ret = append(ret, b)
			}
		}
	}

	return
}

func GetBlackFramesFromVideo(inFileName string) []FFFrameset {

	ffOut := FFBlackdetect(inFileName)

	ffFrames := gmsutils.RemoveDuplicateValues(GrabFFFrameTimes(ffOut))

	ffFrameTimes := gmsutils.Each(ffFrames, func(s string) string { return strings.Split(s, "=")[1] })

	return (ReduceFramesets(ffFrameTimes, 1.0))
}

func FFBlackdetect(inFileName string) string {

	inFileName = gmsutils.SanitizeFilePath(inFileName)

	// call ffprobe blackdetect
	a, err := ffmpeg.Probe(
		"movie="+inFileName+",blackdetect[out0]",
		ffmpeg.KwArgs{
			"f":            "lavfi",
			"show_entries": "tags=lavfi.black_start,lavfi.black_end",
			"of":           "default=nw=1",
		},
	)

	if err != nil {
		panic(err)
	}

	return a
}

func GrabFFFrameTimes(rawffOutput string) []string {

	// filter output for blackdetect tags only
	return gmsutils.Filter(strings.Split(rawffOutput, "\r\n"), func(s string) bool {
		return strings.HasPrefix(s, "TAG:lavfi.black_")
	})
}

func GetFileInfo(inFileName string) FFFileInfo {

	// file basename is filename with no path
	// ie '/home/user/file.mkv' > 'file.mkv'
	fileBasename := filepath.Base(inFileName)

	// file name is filename with no extension
	// ie 'file.mkv' > 'file'
	name := gmsutils.GetFileNameWithoutExtension(fileBasename)

	// get file duration (in seconds) from probe
	fileDuration := GrabFFFileDuration(inFileName)

	// retrieve all blackframe sections of the video
	// and sort by longest to shortest duration
	blackFrames := SortFrames(GetBlackFramesFromVideo(inFileName))

	newFileInfo := FFFileInfo{
		Name:      name,
		Basename:  fileBasename,
		Path:      inFileName,
		Duration:  fileDuration,
		Framesets: blackFrames,
	}

	return newFileInfo
}

func GrabFFFileDuration(inFileName string) float64 {

	// filter output for blackdetect tags only
	durationTag := gmsutils.Find(strings.Split(FFProbeFile(inFileName), "\r\n"), func(s string) bool {
		return strings.HasPrefix(s, "duration=")
	})

	// grab duration
	fileDuration, stErr := strconv.ParseFloat(strings.Split(durationTag, "=")[1], 64)
	if stErr != nil {
		log.Fatal(stErr)
	}

	return fileDuration
}

func SplitVideoByTime(filePath string, tag string, startSeconds float64, endSeconds float64) {

	// use ffprobe 'trim' filter to snip the video,
	// also use setpts to ensure duration is updated
	// finally save to new filename
	ffmpeg.Input(filePath).Trim(ffmpeg.KwArgs{"start": startSeconds, "end": endSeconds}).Filter("setpts", ffmpeg.Args{"PTS-STARTPTS"}).Output(filePath + "-" + tag + "-out.mkv").Run()
}

func FFProbeFile(inFileName string) string {

	// parse file input
	inFileName = gmsutils.SanitizeFilePath(inFileName)

	// call ffprobe show_format
	a, err := ffmpeg.ProbeWithTimeoutExec(
		inFileName,
		0,
		ffmpeg.KwArgs{
			"of":           "default=nw=1",
			"show_entries": "format=duration",
			"loglevel":     "quiet",
		},
	)
	if err != nil {
		panic(err)
	}

	return a
}
