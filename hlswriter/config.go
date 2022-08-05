package hlswriter

import (
	"fmt"
	"github.com/grafov/m3u8"
	uuid "github.com/satori/go.uuid"
)

type Config struct {
	OutputDirectory string `split_words:"true" default:"./hls_output"`
	WindowSize      uint   `split_words:"true" default:"5"`
	WindowCapacity  uint   `split_words:"true" default:"10"`
	MsPerSegment    int64  `split_words:"true" default:"10000"`
	Filename        string `split_words:"true"`
}

func (c Config) New() (*HLSWriter, error) {
	if c.Filename == "" {
		c.Filename = fmt.Sprintf("%s.m3u8", uuid.NewV4())
	}

	playlist, err := m3u8.NewMediaPlaylist(c.WindowSize, c.WindowCapacity)
	if err != nil {
		return &HLSWriter{}, fmt.Errorf("could not create new media playlist: %w", err)
	}

	hlsWriter := HLSWriter{
		outputDirectory: c.OutputDirectory,
		windowSize:      c.WindowSize,
		windowCapacity:  c.WindowCapacity,
		msPerSegment:    c.MsPerSegment,
		filename:        c.Filename,
		playlist:        playlist,
	}

	err = hlsWriter.NewSegmentFile()
	if err != nil {
		return &hlsWriter, fmt.Errorf("could not start segment file: %w", err)
	}

	return &hlsWriter, nil
}
