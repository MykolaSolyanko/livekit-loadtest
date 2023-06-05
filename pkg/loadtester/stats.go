package loadtester

import (
	"time"

	"go.uber.org/atomic"
)

type testerStats struct {
	expectedTracks int
	// trackStats     *trackStats
	stats map[string]*trackStats
	err   error
}

type TrackKind string

const (
	TrackKindVideo TrackKind = "video"
	TrackKindAudio TrackKind = "audio"
	TrackKindData  TrackKind = "data"
)

type trackStats struct {
	trackID      string
	kind         TrackKind
	startedAt    atomic.Time
	packets      atomic.Int64
	bytes        atomic.Int64
	dropped      atomic.Int64
	latency      atomic.Int64
	latencyCount atomic.Int64
}

type summary struct {
	kind         TrackKind
	tracks       int
	packets      int64
	bytes        int64
	latency      int64
	latencyCount int64
	dropped      int64
	elapsed      time.Duration
	errString    string
	errCount     int64
}

func (k TrackKind) String() string {
	return string(k)
}

func getTestSummary(summaries map[string][]*summary) []*summary {
	return []*summary{
		getTestTotalSummary(summaries, TrackKindVideo),
		getTestTotalSummary(summaries, TrackKindData),
	}
}

func getTestTotalSummary(summaries map[string][]*summary, kind TrackKind) *summary {
	s := &summary{}
	for _, testerSummary := range summaries {
		for _, trackSummary := range testerSummary {
			if trackSummary.kind != kind {
				continue
			}

			s.tracks += trackSummary.tracks
			s.packets += trackSummary.packets
			s.kind = trackSummary.kind
			s.bytes += trackSummary.bytes
			s.latency += trackSummary.latency
			s.latencyCount += trackSummary.latencyCount
			s.dropped += trackSummary.dropped
			if trackSummary.elapsed > s.elapsed {
				s.elapsed = trackSummary.elapsed
			}

			s.errCount += trackSummary.errCount
		}
	}

	return s
}

func getTesterSummary(testerStats *testerStats) []*summary {
	return []*summary{
		getTesterTracksSummary(testerStats, TrackKindVideo),
		getTesterTracksSummary(testerStats, TrackKindData),
	}
}

func getTesterTracksSummary(testerStats *testerStats, kind TrackKind) *summary {
	if testerStats == nil {
		return nil
	}

	if testerStats.err != nil {
		return &summary{
			errString: testerStats.err.Error(),
			errCount:  1,
		}
	}

	s := &summary{
		errString: "-",
		kind:      kind,
	}

	for _, trackStats := range testerStats.stats {
		if trackStats.kind != kind {
			continue
		}

		s.tracks++
		s.packets += trackStats.packets.Load()
		s.packets += trackStats.packets.Load()
		s.bytes += trackStats.bytes.Load()
		s.dropped += trackStats.dropped.Load()
		s.latency += trackStats.latency.Load()
		s.latencyCount += trackStats.latencyCount.Load()

		elapsed := time.Since(trackStats.startedAt.Load())
		if elapsed > s.elapsed {
			s.elapsed = elapsed
		}
	}

	return s
}
