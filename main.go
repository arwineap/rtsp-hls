package main

import (
	"fmt"
	"github.com/arwineap/rtsp-hls/rtspviewer"
	"github.com/deepch/vdk/format/rtspv2"
	"go.uber.org/zap"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	conf, err := rtspviewer.LoadConfigFromEnv()
	if err != nil {
		logger.Fatal("could not start rtsp session", zap.Error(err))
	}

	session, err := conf.New()
	if err != nil {
		logger.Fatal("could not connect to rtsp server", zap.Error(err))
	}

	hlsWriter, err := conf.HLS.New()

	if session.CodecData != nil {
		err = hlsWriter.AddCodecs(session.CodecData)
		if err != nil {
			logger.Fatal("could not add initial session codecs", zap.Error(err))
		}
	}

	pingStream := time.NewTimer(15 * time.Second)
	for {
		select {
		case <-pingStream.C:
			panic(fmt.Errorf("stream has no video"))
		case signals := <-session.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				logger.Info("codec update", zap.Any("codec_data", session.CodecData))
				err = hlsWriter.AddCodecs(session.CodecData)
				if err != nil {
					logger.Fatal("unable to add codecs", zap.Error(err))
				}
			case rtspv2.SignalStreamRTPStop:
				logger.Fatal("stream stopped")
			}
		case packetAV := <-session.OutgoingPacketQueue:
			err = hlsWriter.WritePacket(packetAV)
			if err != nil {
				logger.Fatal("could not write packet to hlswriter", zap.Error(err))
			}
			pingStream.Reset(15 * time.Second)
		}
	}
}
