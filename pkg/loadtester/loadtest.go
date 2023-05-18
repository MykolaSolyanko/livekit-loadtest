package loadtester

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
)

type LoadTest struct {
	Params     Params
	trackNames map[string]string
	lock       sync.Mutex
}

type Params struct {
	VideoPublishers int
	AudioPublishers int
	Subscribers     int
	DataPublishers  int
	VideoResolution string
	VideoCodec      string
	Duration        time.Duration
	// number of seconds to spin up per second
	NumPerSecond       float64
	Simulcast          bool
	SimulateSpeakers   bool
	HighQualityViewer  int
	MediumQualityView  int
	LowQualityViewer   int
	DataPacketByteSize int
	DataBitrate        int

	TesterParams
}

func NewLoadTest(params Params) *LoadTest {
	l := &LoadTest{
		Params:     params,
		trackNames: make(map[string]string),
	}

	if l.Params.NumPerSecond == 0 {
		// sane default
		l.Params.NumPerSecond = 5
	}

	if l.Params.NumPerSecond > 10 {
		l.Params.NumPerSecond = 10
	}

	if l.Params.VideoPublishers == 0 && l.Params.AudioPublishers == 0 && l.Params.Subscribers == 0 {
		l.Params.VideoPublishers = 1
		l.Params.Subscribers = 1
	}

	if l.Params.DataPublishers > l.Params.Subscribers {
		l.Params.DataPublishers = l.Params.Subscribers
	}

	if l.Params.DataPacketByteSize == 0 {
		l.Params.DataPacketByteSize = 1024
	}

	if l.Params.DataBitrate == 0 {
		l.Params.DataBitrate = 1024 * 1024 // 1Mbps
	}

	return l
}

func (t *LoadTest) Run(ctx context.Context) error {
	stats, err := t.run(ctx, t.Params)
	if err != nil {
		return err
	}

	summaries := make(map[string]map[string]*summary)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for room, roomStats := range stats {
		fmt.Fprintf(w, "\nStatistics for room %s\n", room)

		summaries[room] = make(map[string]*summary)
		for subName, testerStats := range roomStats {
			summaries[room][subName] = getTesterSummary(testerStats)

			if testerStats.trackStats == nil {
				continue
			}

			_, _ = fmt.Fprintf(w, "\n%s\t| Track\t| Kind\t| Pkts\t| Bitrate\t| Latency\t| Dropped\t| Data Pkts\t| Data Bitrate\t| Latency\n", subName)

			latency, dataChannelLatency, dropped := formatStrings(
				testerStats.trackStats.packets.Load(), testerStats.trackStats.latency.Load(),
				testerStats.trackStats.latencyCount.Load(), testerStats.trackStats.dataChannelLatency.Load(),
				testerStats.trackStats.dataChannelLatencyCount.Load(), testerStats.trackStats.dropped.Load())

			trackName := t.trackNames[testerStats.trackStats.trackID]

			_, _ = fmt.Fprintf(w, "\t| %s %s\t| %s\t| %d\t| %s\t| %s\t| %s\t| %d\t| %s\t| %s\n",
				trackName, testerStats.trackStats.trackID, testerStats.trackStats.kind, testerStats.trackStats.packets.Load(),
				formatBitrate(testerStats.trackStats.bytes.Load(), time.Since(testerStats.trackStats.startedAt.Load())), latency, dropped,
				testerStats.trackStats.dataChannelPackets.Load(),
				formatBitrate(testerStats.trackStats.dataChannelBytes.Load(), time.Since(testerStats.trackStats.dataChannelStartedAt.Load())), dataChannelLatency)

			_ = w.Flush()
		}
	}

	if len(summaries) == 0 {
		return nil
	}

	// summary
	for name, subSummary := range summaries {
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		fmt.Fprintf(w, "\nSummary for room %s\n", name)
		_, _ = fmt.Fprint(w, "\nSummary\t| Tester\t| Bitrate\t| Latency\t| Total Dropped\t| Data Bitrate\t| Latency\t| Error\n")

		for subName, summary := range subSummary {
			sLatency, sDataChannelLatency, sDropped := formatStrings(
				summary.packets, summary.latency, summary.latencyCount, summary.dataChannelLatency,
				summary.dataChannelLatencyCount, summary.dropped)

			sBitrate := formatBitrate(summary.bytes, summary.elapsed)
			sChannelBitrate := formatBitrate(summary.channelBytes, summary.channelElapsed)
			_, _ = fmt.Fprintf(w, "\t| %s\t| %s\t| %s\t| %s\t| %s\t| %s\t| %s\n",
				subName, sBitrate, sLatency, sDropped, sChannelBitrate, sDataChannelLatency, summary.errString)
		}

		s := getTestSummary(subSummary)
		sLatency, sDataChannelLatency, sDropped := formatStrings(
			s.packets, s.latency, s.latencyCount, s.dataChannelLatency, s.dataChannelLatencyCount, s.dropped)
		// avg bitrate per sub
		sBitrate := fmt.Sprintf("%s (%s avg)",
			formatBitrate(s.bytes, s.elapsed),
			formatBitrate(s.bytes/int64(len(summaries)), s.elapsed),
		)
		// avg data channel bitrate per sub
		sChannelBitrate := fmt.Sprintf("%s (%s avg)",
			formatBitrate(s.channelBytes, s.channelElapsed),
			formatBitrate(s.channelBytes/int64(len(summaries)), s.channelElapsed),
		)
		_, _ = fmt.Fprintf(w, "\t| %s\t| %s\t| %s\t| %s\t| %s\t| %s\t| %d\n",
			"Total", sBitrate, sLatency, sDropped, sChannelBitrate, sDataChannelLatency, s.errCount)

		_ = w.Flush()
	}

	_ = w.Flush()
	return nil
}

func (t *LoadTest) GetResolutions() []string {
	resolutions := strings.Split(t.Params.VideoResolution, " ")

	if len(resolutions) < t.Params.VideoPublishers {
		for i := len(resolutions); i < t.Params.VideoPublishers; i++ {
			resolutions = append(resolutions, "1080p")
		}
	}

	return resolutions
}

func (t *LoadTest) run(ctx context.Context, params Params) (map[string]map[string]*testerStats, error) {
	params.IdentityPrefix = randStringRunes(5)

	expectedTracks := params.VideoPublishers + params.AudioPublishers

	var participantStrings []string
	if params.VideoPublishers > 0 {
		participantStrings = append(participantStrings, fmt.Sprintf("%d video publishers", params.VideoPublishers))
	}
	if params.AudioPublishers > 0 {
		participantStrings = append(participantStrings, fmt.Sprintf("%d audio publishers", params.AudioPublishers))
	}
	if params.Subscribers > 0 {
		participantStrings = append(participantStrings, fmt.Sprintf("%d subscribers", params.Subscribers*expectedTracks))
	}

	if params.DataPublishers > 0 {
		participantStrings = append(participantStrings, fmt.Sprintf("%d data publishers", params.DataPublishers))
	}

	fmt.Printf("Starting load test with %s\n", strings.Join(participantStrings, ", "))

	var publishers, testers []*LoadTester
	group, _ := errgroup.WithContext(ctx)
	startedAt := time.Now()
	numStarted := float64(0)
	errs := syncmap.Map{}
	maxPublishers := params.VideoPublishers
	resolutions := t.GetResolutions()

	splitVideoQualityViewers(
		&params.HighQualityViewer, &params.MediumQualityView, &params.LowQualityViewer,
		params.Subscribers, params.Simulcast)

	ready := make(chan struct{})

	for i := 0; i < maxPublishers; i++ {
		room := fmt.Sprintf("%s_%d", params.Room, i)
		testerPubParams := params.TesterParams
		testerPubParams.Sequence = i
		testerPubParams.IdentityPrefix += fmt.Sprintf("_pub%s", room)
		testerPubParams.name = fmt.Sprintf("Pub %d", i)
		testerPubParams.Room = room
		tester := NewLoadTester(testerPubParams, livekit.VideoQuality_HIGH)

		publishers = append(publishers, tester)
		resolution := resolutions[i]

		group.Go(func() error {
			if err := tester.Start(); err != nil {
				fmt.Println(errors.Wrapf(err, "could not connect %s", testerPubParams.name))
				errs.Store(testerPubParams.name, err)
				return nil
			}

			var video string
			var err error
			if params.Simulcast {
				video, err = tester.PublishSimulcastTrack("video-simulcast", resolution, params.VideoCodec)
			} else {
				video, err = tester.PublishVideoTrack("video", resolution, params.VideoCodec)
			}
			if err != nil {
				errs.Store(testerPubParams.name, err)
				return nil
			}
			t.lock.Lock()
			t.trackNames[video] = fmt.Sprintf("%dV", testerPubParams.Sequence)
			t.lock.Unlock()

			return nil
		})
		numStarted++

		high := params.HighQualityViewer
		medium := params.MediumQualityView
		low := params.LowQualityViewer

		var dataPublisher int = 0

		for j := 0; j < params.Subscribers; j++ {
			testerSubParams := params.TesterParams
			testerSubParams.Sequence = j
			testerSubParams.expectedTracks = expectedTracks
			testerSubParams.Subscribe = true
			testerSubParams.Resolution = resolution
			testerSubParams.IdentityPrefix += fmt.Sprintf("_sub%s", room)
			testerSubParams.Room = room
			testerSubParams.name = fmt.Sprintf("Sub %d in %s", j, room)

			quality := livekit.VideoQuality_HIGH

			if high > 0 {
				quality = livekit.VideoQuality_HIGH
				high--
			} else if medium > 0 {
				quality = livekit.VideoQuality_MEDIUM
				medium--
			} else if low > 0 {
				quality = livekit.VideoQuality_LOW
				low--
			}

			tester := NewLoadTester(testerSubParams, quality)
			testers = append(testers, tester)

			group.Go(func() error {
				if err := tester.Start(); err != nil {
					fmt.Println(errors.Wrapf(err, "could not connect %s", testerSubParams.name))
					errs.Store(testerSubParams.name, err)
					return nil
				}

				if dataPublisher >= params.DataPublishers {
					return nil
				}

				dataPublisher++

				if err := tester.PublishData(
					params.DataPacketByteSize, params.DataBitrate, livekit.DataPacket_RELIABLE, ready); err != nil {
					errs.Store(testerSubParams.name, err)
					return nil
				}

				return nil
			})
			numStarted++
		}
	}

	// throttle pace of join events
	for {
		secondsElapsed := float64(time.Since(startedAt)) / float64(time.Second)
		startRate := numStarted / secondsElapsed
		if err := ctx.Err(); err != nil {
			close(ready)

			return nil, err
		}
		if startRate > params.NumPerSecond {
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	var speakerSim *SpeakerSimulator
	if len(publishers) > 0 && t.Params.SimulateSpeakers {
		speakerSim = NewSpeakerSimulator(SpeakerSimulatorParams{
			Testers: publishers,
		})
		speakerSim.Start()
	}

	if err := group.Wait(); err != nil {
		close(ready)

		return nil, err
	}

	duration := params.Duration
	if duration == 0 {
		// a really long time
		duration = 1000 * time.Hour
	}
	fmt.Printf("Finished connecting to room, waiting %s\n", duration.String())
	close(ready)

	select {
	case <-ctx.Done():
		// canceled
	case <-time.After(duration):
		// finished
	}

	if speakerSim != nil {
		speakerSim.Stop()
	}

	stats := make(map[string]map[string]*testerStats)
	for _, t := range testers {
		t.Stop()
		if stats[t.params.Room] == nil {
			stats[t.params.Room] = make(map[string]*testerStats)
		}
		stats[t.params.Room][t.params.name] = t.getStats()
		if e, _ := errs.Load(t.params.name); e != nil {
			stats[t.params.Room][t.params.name].err = e.(error)
		}
	}

	return stats, nil
}

func splitVideoQualityViewers(high, medium, low *int, subscribers int, simulcasting bool) {
	if !simulcasting {
		*high = subscribers
		return
	}

	if *high > subscribers {
		*high = subscribers
		return
	}

	if *high == 0 && *medium == 0 && *low == 0 {
		*high = subscribers
		return
	}

	if *high+*medium > subscribers {
		*medium = subscribers - *high
		return
	}

	if *high+*medium+*low > subscribers {
		*low = subscribers - *high - *medium
		return
	}

	if *high+*medium+*low < subscribers {
		*high += subscribers - *high - *medium - *low
	}
}
