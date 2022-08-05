package rtspviewer

import (
	"fmt"
	"github.com/arwineap/rtsp-hls/hlswriter"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	URL              string        `required:"true"`
	DialTimeout      time.Duration `default:"3s"`
	ReadWriteTimeout time.Duration `default:"3s"`
	Debug            bool          `default:"false"`

	HLS hlswriter.Config
}

func LoadConfigFromEnv() (Config, error) {
	var conf Config
	err := envconfig.Process("APP", &conf)
	return conf, err
}

func (c Config) New() (Client, error) {
	session, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              c.URL,
		DisableAudio:     true,
		DialTimeout:      c.DialTimeout,
		ReadWriteTimeout: c.ReadWriteTimeout,
		Debug:            c.Debug,
	})
	if err != nil {
		return Client{}, fmt.Errorf("could not open rtsp session: %w", err)
	}

	return Client{RTSPClient: session}, err

}
