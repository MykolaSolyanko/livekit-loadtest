package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-cli/pkg/loadtester"
	"github.com/livekit/protocol/logger"
	lksdk "github.com/livekit/server-sdk-go"
)

var LoadTestCommands = []*cli.Command{
	{
		Name:     "load-test",
		Usage:    "Run load tests against LiveKit with simulated publishers & subscribers",
		Category: "Simulate",
		Action:   loadTest,
		Flags: withDefaultFlags(
			&cli.StringFlag{
				Name:  "room",
				Usage: "name of the room (default to random name)",
			},
			&cli.DurationFlag{
				Name:  "duration",
				Usage: "duration to run, 1m, 1h (by default will run until canceled)",
				Value: 0,
			},
			&cli.IntFlag{
				Name:    "video-publishers",
				Aliases: []string{"publishers"},
				Usage:   "number of participants that would publish video tracks",
			},
			&cli.IntFlag{
				Name:  "audio-publishers",
				Usage: "number of participants that would publish audio tracks",
			},
			&cli.IntFlag{
				Name:  "data-publishers",
				Usage: "number of participants that would publish data packets",
			},
			&cli.IntFlag{
				Name:  "data-packet-bytes",
				Usage: "size of data packet in bytes to publish",
				Value: 1024,
			},
			&cli.IntFlag{
				Name:  "data-bitrate",
				Usage: "bitrate in kbps of data channel to publish",
				Value: 1024,
			},
			&cli.IntFlag{
				Name:  "subscribers",
				Usage: "number of participants that would subscribe to tracks",
			},
			&cli.IntFlag{
				Name:  "high",
				Usage: "number of participants that would subscribe to high quality tracks.",
			},
			&cli.IntFlag{
				Name:  "medium",
				Usage: "number of participants that would subscribe to medium quality tracks.",
			},
			&cli.IntFlag{
				Name:  "low",
				Usage: "number of participants that would subscribe to low quality tracks.",
			},
			&cli.StringFlag{
				Name:  "identity-prefix",
				Usage: "identity prefix of tester participants (defaults to a random prefix)",
			},
			&cli.StringFlag{
				Name:  "video-resolution",
				Usage: "resolution of video to publish. valid values are: 360p, 720p, 1080p, 1440p (default: 1080p)",
				Value: "1080p",
			},
			&cli.StringFlag{
				Name:  "video-codec",
				Usage: "h264 or vp8, both will be used when unset",
			},
			&cli.Float64Flag{
				Name:  "num-per-second",
				Usage: "number of testers to start every second",
				Value: 5,
			},
			&cli.StringFlag{
				Name:  "layout",
				Usage: "layout to simulate, choose from speaker, 3x3, 4x4, 5x5",
				Value: "speaker",
			},
			&cli.BoolFlag{
				Name:  "no-simulcast",
				Usage: "disables simulcast publishing (simulcast is enabled by default)",
			},
			&cli.BoolFlag{
				Name:  "simulate-speakers",
				Usage: "fire random speaker events to simulate speaker changes",
			},
			&cli.BoolFlag{
				Name:   "run-all",
				Usage:  "runs set list of load test cases",
				Hidden: true,
			},
		),
	},
}

func loadTest(cCtx *cli.Context) error {
	pc, err := loadProjectDetails(cCtx)
	if err != nil {
		return err
	}

	if !cCtx.Bool("verbose") {
		lksdk.SetLogger(logger.LogRLogger(logr.Discard()))
	}
	_ = raiseULimit()

	ctx, cancel := context.WithCancel(cCtx.Context)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-done
		cancel()
	}()

	params := loadtester.Params{
		VideoResolution:   cCtx.String("video-resolution"),
		VideoCodec:        cCtx.String("video-codec"),
		Duration:          cCtx.Duration("duration"),
		NumPerSecond:      cCtx.Float64("num-per-second"),
		Simulcast:         !cCtx.Bool("no-simulcast"),
		SimulateSpeakers:  cCtx.Bool("simulate-speakers"),
		HighQualityViewer: cCtx.Int("high"),
		MediumQualityView: cCtx.Int("medium"),
		LowQualityViewer:  cCtx.Int("low"),
		TesterParams: loadtester.TesterParams{
			URL:            pc.URL,
			APIKey:         pc.APIKey,
			APISecret:      pc.APISecret,
			Room:           "load-test",
			IdentityPrefix: cCtx.String("identity-prefix"),
		},
	}

	params.VideoPublishers = cCtx.Int("video-publishers")
	params.AudioPublishers = cCtx.Int("audio-publishers")
	params.DataPublishers = cCtx.Int("data-publishers")
	params.Subscribers = cCtx.Int("subscribers")
	params.DataPacketByteSize = cCtx.Int("data-packet-bytes")
	params.DataBitrate = cCtx.Int("data-bitrate") * 1024

	test := loadtester.NewLoadTest(params)
	return test.Run(ctx)
}
