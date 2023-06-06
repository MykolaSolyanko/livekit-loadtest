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
	quality    livekit.VideoQuality
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

func (v *videoSpec) ToVideoLayer() *livekit.VideoLayer {
	return &livekit.VideoLayer{
		Quality: v.quality,
		Height:  uint32(v.height),
		Width:   uint32(v.width),
		Bitrate: v.bitrate(),
	}
}

func (v *videoSpec) bitrate() uint32 {
	return uint32(v.kbps * 1000)
}

func prepareVideoSpecs(prefix string, resolution string, codec string, bitrate ...int) videoSpecParam {
	videoFps := 24

	ratios, ok := resolutions[resolution]
	if !ok {
		panic(fmt.Sprintf("unsupported resolution: %s", resolution))
	}

	if len(bitrate) != len(bitrate) {
		panic("bitrate must match number of ratios")
	}

	specs := make([]*videoSpec, len(ratios))
	for i, bitrate := range bitrate {
		specs[i] = &videoSpec{
			prefix:     prefix,
			codec:      codec,
			kbps:       bitrate,
			fps:        videoFps,
			resolution: resolution,
			height:     ratios[i].Height,
			width:      ratios[i].Width,
			quality:    ratios[i].Quality,
		}
	}

	return videoSpecParam{
		specs:      specs,
		resolution: resolution,
		codec:      codec,
	}
}

type Ratio struct {
	Width   int
	Height  int
	Quality livekit.VideoQuality
}

type videoSpecParam struct {
	codec      string
	resolution string
	specs      []*videoSpec
}

var (
	//go:embed resources
	res embed.FS

	videoSpecs  []videoSpecParam
	videoIndex  atomic.Int64
	audioIndex  atomic.Int64
	resolutions map[string][]Ratio
)

func prepareResolutions() {
	resolutions = make(map[string][]Ratio)
	resolutions["1440p"] = []Ratio{
		{Width: 2560, Height: 1440, Quality: livekit.VideoQuality_HIGH},
		{Width: 2048, Height: 1152, Quality: livekit.VideoQuality_MEDIUM},
		{Width: 1024, Height: 576, Quality: livekit.VideoQuality_LOW},
	}

	resolutions["1080p"] = []Ratio{
		{Width: 1920, Height: 1080, Quality: livekit.VideoQuality_HIGH},
		{Width: 800, Height: 450, Quality: livekit.VideoQuality_MEDIUM},
		{Width: 640, Height: 360, Quality: livekit.VideoQuality_LOW},
	}
	resolutions["720p"] = []Ratio{
		{Width: 1280, Height: 720, Quality: livekit.VideoQuality_HIGH},
		{Width: 800, Height: 450, Quality: livekit.VideoQuality_MEDIUM},
		{Width: 640, Height: 360, Quality: livekit.VideoQuality_LOW},
	}

	resolutions["360p"] = []Ratio{
		{Width: 640, Height: 360, Quality: livekit.VideoQuality_HIGH},
		{Width: 640, Height: 360, Quality: livekit.VideoQuality_MEDIUM},
		{Width: 640, Height: 360, Quality: livekit.VideoQuality_LOW},
	}
}

func GetVideoResolution(resolution string) []Ratio {
	return resolutions[resolution]
}

func init() {
	prepareResolutions()

	videoSpecs = []videoSpecParam{
		prepareVideoSpecs("butterfly", "360p", h264Codec, 460, 460, 460),
		prepareVideoSpecs("butterfly", "720p", h264Codec, 1800, 720, 460),
		prepareVideoSpecs("butterfly", "1080p", h264Codec, 4100, 720, 460),
		prepareVideoSpecs("butterfly", "1440p", h264Codec, 7300, 4800, 1200),
	}
}

func getVideoSpecs(videoCodec string, resolution string) []*videoSpec {
	if videoCodec == "" {
		videoCodec = h264Codec
	}

	if resolution == "" {
		resolution = "1080p"
	}

	for _, spec := range videoSpecs {
		if spec.codec == videoCodec && spec.resolution == resolution {
			return spec.specs
		}
	}

	return nil
}

func CreateVideoLoopers(resolution string, codecFilter string, simulcast bool) ([]VideoLooper, error) {
	specs := getVideoSpecs(codecFilter, resolution)
	if specs == nil {
		return nil, fmt.Errorf("could not find video spec for %s %s", codecFilter, resolution)
	}

	var loopers []VideoLooper

	if !simulcast {
		specs = specs[:1]
	}

	for _, spec := range specs {
		f, err := res.Open(spec.Name())
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var looper VideoLooper
		if spec.codec == h264Codec {
			if looper, err = NewH264VideoLooper(f, spec); err != nil {
				return nil, err
			}
		} else if spec.codec == vp8Codec {
			if looper, err = NewVP8VideoLooper(f, spec); err != nil {
				return nil, err
			}
		}

		loopers = append(loopers, looper)
	}

	return loopers, nil
}

func CreateAudioLooper() (*OpusAudioLooper, error) {
	f, err := res.Open("resources/change-ken.ogg")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewOpusAudioLooper(f)
}
