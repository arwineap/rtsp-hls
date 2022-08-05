package hlswriter

import (
	"fmt"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/ts"
	"github.com/grafov/m3u8"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HLSWriter struct {
	outputDirectory string
	windowSize      uint
	windowCapacity  uint
	msPerSegment    int64
	filename        string

	playlist       *m3u8.MediaPlaylist
	lastPacketTime time.Time
	lastKeyFrame   av.Packet
	segmentNumber  int64
	codecs         codecDataStruct
	tsMuxer        *ts.Muxer
	segmentStart   time.Time
	outFile        *os.File
}

func (h *HLSWriter) GetSegmentFilename() string {
	return fmt.Sprintf("%04d.ts", h.segmentNumber)
}

func (h *HLSWriter) NewSegmentFile() error {
	h.segmentNumber++

	segmentName := h.GetSegmentFilename()
	segmentPath := filepath.Join(h.outputDirectory, segmentName)
	var err error
	h.outFile, err = os.Create(segmentPath)
	if err != nil {
		return fmt.Errorf("could not create new segment file: %w", err)
	}

	h.tsMuxer = ts.NewMuxer(h.outFile)

	// Write header
	codecData := h.codecs.get()
	err = h.tsMuxer.WriteHeader(codecData)
	if err != nil {
		return fmt.Errorf("could not write header to new segment file: %w", err)
	}

	h.segmentStart = time.Now()

	return nil
}

func (h *HLSWriter) EndSegmentFile() error {
	err := h.tsMuxer.WriteTrailer()
	if err != nil {
		return fmt.Errorf("unable to write trailer for segment: %w", err)
	}

	err = h.outFile.Close()
	if err != nil {
		return fmt.Errorf("unable to close segment file: %w", err)
	}

	err = h.SlidePlaylist()
	if err != nil {
		return fmt.Errorf("could not update playlist: %w", err)
	}

	err = h.PurgeOutdatedSegments()
	if err != nil {
		return fmt.Errorf("unable to purge outdated segments: %w", err)
	}

	return nil
}

func (h *HLSWriter) PurgeOutdatedSegments() error {
	// create a lookup table for segments
	currentSegments := make(map[string]struct{}, len(h.playlist.Segments))
	for _, segment := range h.playlist.Segments {
		if segment != nil {
			currentSegments[segment.URI] = struct{}{}
		}
	}

	segmentFiles, err := filepath.Glob(filepath.Join(h.outputDirectory, "*.ts"))
	if err != nil {
		return fmt.Errorf("could not glob through ts files in output directory: %w", err)
	}

	for _, segmentFile := range segmentFiles {
		if _, ok := currentSegments[filepath.Base(segmentFile)]; !ok {
			err = os.Remove(segmentFile)
			if err != nil {
				return fmt.Errorf("unable to remove segment file ( %s ): %w", filepath.Base(segmentFile), err)
			}
		}
	}
	return nil
}

func (h *HLSWriter) SlidePlaylist() error {
	h.playlist.Slide(h.GetSegmentFilename(), time.Since(h.segmentStart).Seconds(), "")
	playlistFile, err := os.Create(filepath.Join(h.outputDirectory, h.filename))
	if err != nil {
		return fmt.Errorf("unable to create playlistfile: %w", err)
	}
	_, err = playlistFile.Write(h.playlist.Encode().Bytes())
	if err != nil {
		return fmt.Errorf("could not write playlist file: %w", err)
	}
	err = playlistFile.Close()
	if err != nil {
		return fmt.Errorf("could not close playlist file: %w", err)
	}

	return nil
}

func (h *HLSWriter) WritePacket(pkt *av.Packet) error {
	h.lastPacketTime = time.Now()

	if pkt.IsKeyFrame {
		if time.Since(h.segmentStart).Milliseconds() > h.msPerSegment {
			err := h.EndSegmentFile()
			if err != nil {
				return fmt.Errorf("could not end segment file: %w", err)
			}
			err = h.NewSegmentFile()
			if err != nil {
				return fmt.Errorf("unable to create new segment file: %w", err)
			}
		}
	}

	return h.tsMuxer.WritePacket(*pkt)
}

func (h *HLSWriter) AddCodecs(newCodecs []av.CodecData) error {
	h.codecs.add(newCodecs)

	err := h.tsMuxer.WriteHeader(newCodecs)
	if err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	return nil
}

type codecDataStruct struct {
	codecs []av.CodecData

	sync.Mutex
}

func (c *codecDataStruct) add(newCodecs []av.CodecData) {
	defer c.Unlock()
	c.Lock()

	c.codecs = append(c.codecs, newCodecs...)
}

func (c *codecDataStruct) get() []av.CodecData {
	defer c.Unlock()
	c.Lock()

	return c.codecs
}
