package provider

import (
	"embed"
	"fmt"
	"strconv"

	"go.uber.org/atomic"

	"github.com/livekit/protocol/livekit"
)

const (
	h264Codec = "h264"
	vp8Codec  = "vp8"
)

type videoSpec struct {
	codec      string
	prefix     string
	height     int
	width      int
	kbps       int
	fps        int
	resolution string
}

func (v *videoSpec) Name() string {
	ext := "h264"
	if v.codec == vp8Codec {
		ext = "ivf"
	}
	size := strconv.Itoa(v.height)
	if v.height > v.width {
		size = fmt.Sprintf("p%d", v.width)
	}
	return fmt.Sprintf("resources/%s_%s_%d.%s", v.prefix, size, v.kbps, ext)
}

func (v *videoSpec) ToVideoLayer(quality livekit.VideoQuality) *livekit.VideoLayer {
	return &livekit.VideoLayer{
		Quality: quality,
		Height:  uint32(v.height),
		Width:   uint32(v.width),
		Bitrate: v.bitrate(),
	}
}

func (v *videoSpec) bitrate() uint32 {
	return uint32(v.kbps * 1000)
}

func createSpec(prefix string, resolution string, codec string, bitrate int) *videoSpec {
	videoFps := 24

	var height, width int

	switch resolution {
	case "360p":
		height = 360
		width = 640
	case "720p":
		height = 720
		width = 1280
	case "1080p":
		height = 1080
		width = 1920
	case "1440p":
		height = 1440
		width = 2560
	default:
		height = 1080
		width = 1920
	}

	return &videoSpec{
		prefix:     prefix,
		codec:      codec,
		kbps:       bitrate,
		fps:        videoFps,
		resolution: resolution,
		height:     height,
		width:      width,
	}
}

var (
	//go:embed resources
	res embed.FS

	videoSpecs []*videoSpec
	videoIndex atomic.Int64
	audioNames []string
	audioIndex atomic.Int64
)

func init() {
	videoSpecs = []*videoSpec{
		createSpec("butterfly", "360p", h264Codec, 460),
		createSpec("butterfly", "720p", h264Codec, 1800),
		createSpec("butterfly", "1080p", h264Codec, 4100),
		createSpec("butterfly", "1440p", h264Codec, 7300),
	}
	audioNames = []string{
		"change-amelia",
		"change-benjamin",
		"change-elena",
		"change-clint",
		"change-emma",
		"change-ken",
		"change-sophie",
	}
}

func getVideoSpecs(videoCodec string, resolution string) *videoSpec {
	if videoCodec == "" {
		videoCodec = h264Codec
	}

	if resolution == "" {
		resolution = "1080p"
	}

	for _, spec := range videoSpecs {
		if spec.codec == videoCodec && spec.resolution == resolution {
			return spec
		}
	}

	return nil
}

func CreateVideoLooper(resolution string, codecFilter string) (VideoLooper, error) {
	spec := getVideoSpecs(codecFilter, resolution)
	if spec == nil {
		return nil, fmt.Errorf("could not find video spec for %s %s", codecFilter, resolution)
	}

	var looper VideoLooper

	f, err := res.Open(spec.Name())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if spec.codec == h264Codec {
		if looper, err = NewH264VideoLooper(f, spec); err != nil {
			return nil, err
		}
	} else if spec.codec == vp8Codec {
		if looper, err = NewVP8VideoLooper(f, spec); err != nil {
			return nil, err
		}
	}

	return looper, nil
}

func CreateAudioLooper() (*OpusAudioLooper, error) {
	chosenName := audioNames[int(audioIndex.Load())%len(audioNames)]
	audioIndex.Inc()
	f, err := res.Open(fmt.Sprintf("resources/%s.ogg", chosenName))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewOpusAudioLooper(f)
}
