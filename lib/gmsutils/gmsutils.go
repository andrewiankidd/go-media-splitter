package gmsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/andrewiankidd/go-media-splitter/lib/gmstypes"
)

var (
	Verbose *bool
)

type FFFileInfo = gmstypes.FFFileInfo
type FFFrameset = gmstypes.FFFrameset

func Filter(ss []string, test func(string) bool) (ret []string) {

	// iterate each entry of the input array
	for _, s := range ss {

		// if it passes test function
		if test(s) {

			// append to return object
			ret = append(ret, s)
		}
	}
	return
}

func Find(ss []string, test func(string) bool) (ret string) {

	// iterate each entry of the input array
	for _, s := range ss {

		// if it passes test function
		if test(s) {

			// return object
			ret = s
			return
		}
	}
	return
}

func Each(ss []string, update func(string) string) (ret []string) {

	// iterate each entry of the input array
	for _, s := range ss {

		// add result of update function to return object
		ret = append(ret, update(s))
	}
	return
}

func SanitizeFilePath(inputPath string) string {

	if runtime.GOOS == "windows" {
		// HACK HACK HACK
		// try and make the path work with fftools
		inputPath = strings.Replace(filepath.ToSlash(inputPath), "C:", "", 1)
	}

	return inputPath
}

func RemoveDuplicateValues(stringSlice []string) []string {

	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

// https://gist.github.com/ivanzoid/129460aa08aff72862a534ebe0a9ae30#gistcomment-3733302
func GetFileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

// break the data down into fixed sized 'chunks'
func Chunk(xs []float64, chunkSize int) [][]float64 {
	// https://stackoverflow.com/a/67011816

	if len(xs) == 0 {
		return nil
	}
	divided := make([][]float64, (len(xs)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(xs) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = xs[prev:next]
		prev = next
		i++
	}
	divided[i] = xs[prev:]
	return divided
}

// https://stackoverflow.com/a/55300382
func WalkMatch(root string, pattern string) ([]string, error) {

	var matches []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

func EstimateSweetspots(fInfo FFFileInfo) []FFFrameset {

	episodeCount := EstimateEpisodeCount(fInfo.Name)
	fmt.Printf("estimated episode count: %v\r\n", episodeCount)

	// length of media split by episode count guess
	equalParts := fInfo.Duration / float64(episodeCount)

	// these will be areas you'd expect the media to change 'episode',
	// if episodes were of equal length
	sweetspots := []FFFrameset{}
	for i := 1; i < episodeCount; i++ {

		// 30 second fuzz in either direction
		secondFuzz := float64(30)

		// prepare new frameset
		// pad equal parts x mins in each direction
		// +/- a few seconds
		f := FFFrameset{
			Start: (equalParts * float64(i)) - (secondFuzz / 2), End: (equalParts * float64(i)) + (secondFuzz / 2),
			Duration: (secondFuzz),
		}

		// add to recommended list of sweetspots to split the media
		sweetspots = append(sweetspots, f)
	}

	// check for any potential matches in where we think the file should be split,
	// vs where any black frames have been detected
	estimatedCutpoints := GetOverlappingFramesets(sweetspots, fInfo.Framesets)

	return estimatedCutpoints
}

func EstimateEpisodeCount(filename string) int {

	// take filename, how many 'parts' could it be?
	// ie grab by E**, so
	// S01E1E02 becomes S01(E01)(E02)

	episodeMatcher := regexp.MustCompile("E([0-9]+)")
	matches := episodeMatcher.FindAllStringSubmatch(filename, -1)
	x := matches[len(matches)-1][1]
	value, err := strconv.Atoi(x)
	if err != nil {
		panic(err)
	}

	return value
}

func GetOverlappingFramesets(sweetSpots []FFFrameset, blackFrames []FFFrameset) (ret []FFFrameset) {

	// loop over sweetspots
	for _, s := range sweetSpots {
		if *Verbose {
			fmt.Printf("=> sweetspot: %v\r\n", s)
		}

		// loop over blackframes
		for _, b := range blackFrames {

			if *Verbose {
				fmt.Printf("==> blackframe: %v\r\n", b)
			}

			// check each blackframe for overlap with the current sweetspot
			if b.Start >= s.Start && b.Start <= s.End {

				if *Verbose {
					fmt.Printf("==> Match! (blackframe starts in sweetspot)\r\n")
				}
			} else if b.End <= s.End && b.End >= s.Start {

				if *Verbose {
					fmt.Printf("==> Match! (blackframe ends in sweetspot)\r\n")
				}
			} else {

				if *Verbose {
					fmt.Printf("==> No Match...\r\n")
				}
				continue
			}

			// add to overlap list to return
			ret = append(ret, b)
		}
	}

	// return list
	return ret
}
