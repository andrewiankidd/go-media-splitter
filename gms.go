package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/andrewiankidd/go-media-splitter/lib/gmsfftools"
	"github.com/andrewiankidd/go-media-splitter/lib/gmstypes"
	"github.com/andrewiankidd/go-media-splitter/lib/gmsutils"
)

var (
	inputDirectory *string
	verbose        *bool
)

type FFFileInfo = gmstypes.FFFileInfo
type FFFrameset = gmstypes.FFFrameset

// Basic flag declarations are available for string, integer, and boolean options.
func init() {
	inputDirectory = flag.String("inputDirectory", "testData", "Specify directory to scan for files.")
	verbose = flag.Bool("verbose", false, "verbose output")
	gmsutils.Verbose = verbose
}

func main() {
	// parse command-line flags
	flag.Parse()

	// scan input directory for matching for files
	fmt.Printf("> Scanning directory: %v\n", *inputDirectory)
	files, err := gmsutils.WalkMatch(*inputDirectory, "*.mkv")

	// handle any errors
	if err != nil {
		log.Fatal(err)
	}

	// handle each file found in the given directory
	for _, fileName := range files {
		fmt.Printf("Scanning file: (%v)\n", fileName)

		// get necessary file info from ffprobe
		FFFileInfo := gmsfftools.GetFileInfo(fileName)

		// estimate cutting media points from filename and ffprobe blackframe scan
		sweetspotEst := gmsutils.EstimateSweetspots(FFFileInfo)
		if *verbose {
			fmt.Printf("(%v) Sweetspot(s): %v\r\n", FFFileInfo.Basename, sweetspotEst)
		}

		// assume where ~~ is blackframe in a 19 sec clip
		// XXX~~XXXXXX~~XXXXX
		//      4 5    12 13
		// prepend 0, append total length
		// 0,4 5,12 13,19
		cutpoints := []float64{float64(0)}
		for _, sweetspot := range sweetspotEst {

			fmt.Printf("(%v) Found: %v second gap found at %v - %v\r\n", FFFileInfo.Basename, sweetspot.Duration, sweetspot.Start, sweetspot.End)

			wiggleroom := sweetspot.Start + (sweetspot.Duration / 2)
			cutpoints = append(cutpoints, wiggleroom)
		}
		cutpoints = append(cutpoints, FFFileInfo.Duration)

		chunkedCutpoints := gmsutils.Chunk(cutpoints, 2)
		// loop + splitvideobyframe
		for cutpointIndex, cutpoint := range chunkedCutpoints {
			fmt.Printf("(%v) Cutpoint: %v - %v\r\n", FFFileInfo.Basename, cutpoint[0], cutpoint[1])

			gmsfftools.SplitVideoByTime(FFFileInfo.Path, fmt.Sprintf("%v", cutpointIndex), cutpoint[0], cutpoint[1])
		}
	}

	fmt.Printf("Done!\n")
}
