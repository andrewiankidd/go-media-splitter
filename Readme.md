# Go Media Splitter

A simple tool that attempts to detect and split multi-part media files.

## Table of Contents

- [Go Media Splitter](#go-media-splitter)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Usage](#usage)
  - [How it works](#how-it-works)
  - [More](#more)

## Installation

- clone or download the repo
- open directory in a terminal
- run `go run gms.go`

## Usage
`gms -inputDirectory ~/exampledirectory`

## How it works

`gms` works by analyzing the filenames of media files, and ffinfo analysis for prolonged periods of 'black frames' in the video stream.

 In order for this to work, filenames need to follow a pattern, ie `{n} - {s00e00} - {t}`. Specifically `gms` looks for re pattern `"E([0-9]+)"`

 To run, execute `gms -inputDirectory ~/exampledirectory`, the directory will be scanned and each will be processed as so

  - file is scanned by ffinfo to obtain some info
    - filename, duration, periods of black frames
  - a list of of potential 'cut points' are generated, by analyzing the filename
  - black frame list and cut points are compared
  - any matches are saved to a new file
  
## More

This is an experimental side-project, I hold no license, support, etc - thanks!