package loadtester

import (
	"time"

	"go.uber.org/atomic"

	lksdk "github.com/livekit/server-sdk-go"
)

type testerStats struct {
	expectedTracks int
	trackStats     *trackStats
	err            error
}

type trackStats struct {
	trackID                 string
	kind                    lksdk.TrackKind
	startedAt               atomic.Time
	packets                 atomic.Int64
	bytes                   atomic.Int64
	dataChannelStartedAt    atomic.Time
	dataChannelPackets      atomic.Int64
	dataChannelBytes        atomic.Int64
	dropped                 atomic.Int64
	latency                 atomic.Int64
	latencyCount            atomic.Int64
	dataChannelLatency      atomic.Int64
	dataChannelLatencyCount atomic.Int64
}

type summary struct {
	packets                 int64
	bytes                   int64
	channelBytes            int64
	channelPackets          int64
	latency                 int64
	latencyCount            int64
	dataChannelLatency      int64
	dataChannelLatencyCount int64
	dropped                 int64
	elapsed                 time.Duration
	channelElapsed          time.Duration
	errString               string
	errCount                int64
}

func getTestSummary(summaries map[string]*summary) *summary {
	s := &summary{}
	for _, testerSummary := range summaries {
		s.packets += testerSummary.packets
		s.bytes += testerSummary.bytes
		s.latency += testerSummary.latency
		s.latencyCount += testerSummary.latencyCount
		s.dataChannelLatency += testerSummary.dataChannelLatency
		s.dataChannelLatencyCount += testerSummary.dataChannelLatencyCount
		s.dropped += testerSummary.dropped
		s.channelBytes += testerSummary.channelBytes
		s.channelPackets += testerSummary.channelPackets
		if testerSummary.elapsed > s.elapsed {
			s.elapsed = testerSummary.elapsed
		}

		if testerSummary.channelElapsed > s.channelElapsed {
			s.channelElapsed = testerSummary.channelElapsed
		}

		s.errCount += testerSummary.errCount
	}
	return s
}

func getTesterSummary(testerStats *testerStats) *summary {
	if testerStats.err != nil {
		return &summary{
			errString: testerStats.err.Error(),
			errCount:  1,
		}
	}

	s := &summary{
		packets:                 testerStats.trackStats.packets.Load(),
		bytes:                   testerStats.trackStats.bytes.Load(),
		channelBytes:            testerStats.trackStats.dataChannelBytes.Load(),
		channelPackets:          testerStats.trackStats.dataChannelPackets.Load(),
		dropped:                 testerStats.trackStats.dropped.Load(),
		latency:                 testerStats.trackStats.latency.Load(),
		latencyCount:            testerStats.trackStats.latencyCount.Load(),
		dataChannelLatency:      testerStats.trackStats.dataChannelLatency.Load(),
		dataChannelLatencyCount: testerStats.trackStats.dataChannelLatencyCount.Load(),
		errString:               "-",
	}

	elapsed := time.Since(testerStats.trackStats.startedAt.Load())
	if elapsed > s.elapsed {
		s.elapsed = elapsed
	}

	channelElapsed := time.Since(testerStats.trackStats.dataChannelStartedAt.Load())
	if channelElapsed > s.channelElapsed {
		s.channelElapsed = channelElapsed
	}

	return s
}
