package loadtester

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

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
	VideoResolution string
	VideoCodec      string
	Duration        time.Duration
	// number of seconds to spin up per second
	NumPerSecond     float64
	Simulcast        bool
	SimulateSpeakers bool

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
	return l
}

func (t *LoadTest) Run(ctx context.Context) error {
	stats, err := t.run(ctx, t.Params)
	if err != nil {
		return err
	}

	// tester results
	summaries := make(map[string]*summary)
	names := make([]string, 0, len(stats))
	for name := range stats {
		if strings.HasPrefix(name, "Pub") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		testerStats := stats[name]
		summaries[name] = getTesterSummary(testerStats)

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintf(w, "\n%s\t| Track\t| Kind\t| Pkts\t| Bitrate\t| Dropped\n", name)
		trackStatsSlice := make([]*trackStats, 0, len(testerStats.trackStats))
		for _, ts := range testerStats.trackStats {
			trackStatsSlice = append(trackStatsSlice, ts)
		}
		sort.Slice(trackStatsSlice, func(i, j int) bool {
			nameI := t.trackNames[trackStatsSlice[i].trackID]
			nameJ := t.trackNames[trackStatsSlice[j].trackID]
			return strings.Compare(nameI, nameJ) < 0
		})
		for _, trackStats := range trackStatsSlice {
			dropped := formatStrings(
				trackStats.packets.Load(), trackStats.dropped.Load())

			trackName := t.trackNames[trackStats.trackID]
			_, _ = fmt.Fprintf(w, "\t| %s %s\t| %s\t| %d\t| %s\t| %s\n",
				trackName, trackStats.trackID, trackStats.kind, trackStats.packets.Load(),
				formatBitrate(trackStats.bytes.Load(), time.Since(trackStats.startedAt.Load())), dropped)
		}
		_ = w.Flush()
	}

	if len(summaries) == 0 {
		return nil
	}

	// summary
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprint(w, "\nSummary\t| Tester\t| Tracks\t| Bitrate\t| Total Dropped\t| Error\n")

	for _, name := range names {
		s := summaries[name]
		sDropped := formatStrings(s.packets, s.dropped)
		sBitrate := formatBitrate(s.bytes, s.elapsed)
		_, _ = fmt.Fprintf(w, "\t| %s\t| %d/%d\t| %s\t| %s\t| %s\n",
			name, s.tracks, s.expected, sBitrate, sDropped, s.errString)
	}

	s := getTestSummary(summaries)
	sDropped := formatStrings(s.packets, s.dropped)
	// avg bitrate per sub
	sBitrate := fmt.Sprintf("%s (%s avg)",
		formatBitrate(s.bytes, s.elapsed),
		formatBitrate(s.bytes/int64(len(summaries)), s.elapsed),
	)
	_, _ = fmt.Fprintf(w, "\t| %s\t| %d/%d\t| %s\t| %s\t| %d\n",
		"Total", s.tracks, s.expected, sBitrate, sDropped, s.errCount)

	_ = w.Flush()
	return nil
}

func (t *LoadTest) RunSuite(ctx context.Context) error {
	cases := []*struct {
		publishers  int
		subscribers int
		video       bool

		tracks  int64
		latency time.Duration
		dropped float64
	}{
		{publishers: 10, subscribers: 10, video: false},
		{publishers: 10, subscribers: 100, video: false},
		{publishers: 10, subscribers: 500, video: false},
		{publishers: 10, subscribers: 1000, video: false},
		{publishers: 50, subscribers: 50, video: false},
		{publishers: 100, subscribers: 50, video: false},

		{publishers: 10, subscribers: 10, video: true},
		{publishers: 10, subscribers: 100, video: true},
		{publishers: 10, subscribers: 500, video: true},
		{publishers: 1, subscribers: 100, video: true},
		{publishers: 1, subscribers: 1000, video: true},
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprint(w, "\nPubs\t| Subs\t| Tracks\t| Audio\t| Video\t| Packet loss\t| Errors\n")

	for _, c := range cases {
		caseParams := t.Params
		videoString := "Yes"
		if c.video {
			caseParams.VideoPublishers = c.publishers
		} else {
			caseParams.AudioPublishers = c.publishers
			videoString = "No"
		}
		caseParams.Subscribers = c.subscribers
		caseParams.Simulcast = true
		if caseParams.Duration == 0 {
			caseParams.Duration = 15 * time.Second
		}
		fmt.Printf("\nRunning test: %d pub, %d sub, video: %s\n", c.publishers, c.subscribers, videoString)

		stats, err := t.run(ctx, caseParams)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return err
		}

		var tracks, packets, dropped, errCount int64
		for _, testerStats := range stats {
			for _, trackStats := range testerStats.trackStats {
				tracks++
				packets += trackStats.packets.Load()
				dropped += trackStats.dropped.Load()
			}
			if testerStats.err != nil {
				errCount++
			}
		}
		_, _ = fmt.Fprintf(w, "%d\t| %d\t| %d\t| Yes\t| %s\t| %.3f%%| %d\t\n",
			c.publishers, c.subscribers, tracks, videoString, 100*float64(dropped)/float64(dropped+packets), errCount)
	}

	_ = w.Flush()
	return nil
}

func (t *LoadTest) run(ctx context.Context, params Params) (map[string]*testerStats, error) {
	// params.Room = fmt.Sprintf("testroom%d", rand.Int31n(1000))
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

	fmt.Printf("Starting load test with %s\n", strings.Join(participantStrings, ", "))

	var publishers, testers []*LoadTester
	group, _ := errgroup.WithContext(ctx)
	startedAt := time.Now()
	numStarted := float64(0)
	errs := syncmap.Map{}
	maxPublishers := params.VideoPublishers

	for i := 0; i < maxPublishers; i++ {
		room := fmt.Sprintf("%s_%d", params.Room, i)
		testerPubParams := params.TesterParams
		testerPubParams.Sequence = i
		testerPubParams.IdentityPrefix += fmt.Sprintf("_pub%s", room)
		testerPubParams.name = fmt.Sprintf("Pub %d", i)
		testerPubParams.Room = room
		tester := NewLoadTester(testerPubParams)
		testers = append(testers, tester)

		publishers = append(publishers, tester)

		group.Go(func() error {
			if err := tester.Start(); err != nil {
				fmt.Println(errors.Wrapf(err, "could not connect %s", testerPubParams.name))
				errs.Store(testerPubParams.name, err)
				return nil
			}

			var video string
			var err error
			if params.Simulcast {
				video, err = tester.PublishSimulcastTrack("video-simulcast", params.VideoResolution, params.VideoCodec)
			} else {
				video, err = tester.PublishVideoTrack("video", params.VideoResolution, params.VideoCodec)
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

		for j := 0; j < params.Subscribers; j++ {
			testerSubParams := params.TesterParams
			testerSubParams.Sequence = j
			testerSubParams.expectedTracks = expectedTracks
			testerSubParams.Subscribe = true
			testerSubParams.IdentityPrefix += fmt.Sprintf("_sub%s", room)
			testerSubParams.Room = room
			testerSubParams.name = fmt.Sprintf("Sub %d in %s", j, room)

			tester := NewLoadTester(testerSubParams)
			testers = append(testers, tester)

			group.Go(func() error {
				if err := tester.Start(); err != nil {
					fmt.Println(errors.Wrapf(err, "could not connect %s", testerSubParams.name))
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
		return nil, err
	}

	duration := params.Duration
	if duration == 0 {
		// a really long time
		duration = 1000 * time.Hour
	}
	fmt.Printf("Finished connecting to room, waiting %s\n", duration.String())

	select {
	case <-ctx.Done():
		// canceled
	case <-time.After(duration):
		// finished
	}

	if speakerSim != nil {
		speakerSim.Stop()
	}

	stats := make(map[string]*testerStats)
	for _, t := range testers {
		t.Stop()
		stats[t.params.name] = t.getStats()
		if e, _ := errs.Load(t.params.name); e != nil {
			stats[t.params.name].err = e.(error)
		}
	}

	return stats, nil
}
