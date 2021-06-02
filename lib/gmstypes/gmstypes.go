package gmstypes

// FFFileInfo type for file metadata
type FFFileInfo struct {
	Name, Basename, Path string
	Duration             float64
	Framesets            []FFFrameset
}

// frame type to store black frame data
type FFFrameset struct {
	Start, End, Duration float64
}
