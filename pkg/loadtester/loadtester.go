package loadtester

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"go.uber.org/atomic"

	provider2 "github.com/livekit/livekit-cli/pkg/provider"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
)

type LoadTester struct {
	params TesterParams

	lock    sync.Mutex
	room    *lksdk.Room
	running atomic.Bool
	// participant ID => quality
	trackQualities    map[string]livekit.VideoQuality
	quality           livekit.VideoQuality
	dataPublishing    atomic.Bool
	dataChannelLoaded atomic.Bool
	stats             atomic.Value
	// stats             *trackStats
}

type TesterParams struct {
	URL            string
	APIKey         string
	APISecret      string
	Room           string
	IdentityPrefix string
	Resolution     string
	// true to subscribe to all published tracks
	Subscribe bool

	name           string
	Sequence       int
	expectedTracks int
}

func NewLoadTester(params TesterParams, quality livekit.VideoQuality) *LoadTester {
	return &LoadTester{
		params:         params,
		quality:        quality,
		trackQualities: make(map[string]livekit.VideoQuality),
	}
}

func (t *LoadTester) Start() error {
	if t.IsRunning() {
		return nil
	}

	identity := fmt.Sprintf("%s_%d", t.params.IdentityPrefix, t.params.Sequence)
	t.room = lksdk.CreateRoom(&lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: t.onTrackSubscribed,
			OnTrackSubscriptionFailed: func(sid string, rp *lksdk.RemoteParticipant) {
				fmt.Printf("track subscription failed, lp:%v, sid:%v, rp:%v/%v\n", identity, sid, rp.Identity(), rp.SID())
			},
			OnTrackPublished: t.onTrackPublished,
			OnDataReceived:   t.onDataReceived,
		},
	})
	var err error
	// make up to 10 reconnect attempts
	for i := 0; i < 10; i++ {
		err = t.room.Join(t.params.URL, lksdk.ConnectInfo{
			APIKey:              t.params.APIKey,
			APISecret:           t.params.APISecret,
			RoomName:            t.params.Room,
			ParticipantIdentity: identity,
		}, lksdk.WithAutoSubscribe(false))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}

	t.running.Store(true)
	for _, p := range t.room.GetParticipants() {
		for _, pub := range p.Tracks() {
			if remotePub, ok := pub.(*lksdk.RemoteTrackPublication); ok {
				t.onTrackPublished(remotePub, p)
			}
		}
	}

	return nil
}

func (t *LoadTester) IsRunning() bool {
	return t.running.Load()
}

func (t *LoadTester) PublishAudioTrack(name string) (string, error) {
	if !t.IsRunning() {
		return "", nil
	}

	fmt.Println("publishing audio track -", t.room.LocalParticipant.Identity())
	audioLooper, err := provider2.CreateAudioLooper()
	if err != nil {
		return "", err
	}
	track, err := lksdk.NewLocalSampleTrack(audioLooper.Codec())
	if err != nil {
		return "", err
	}
	if err := track.StartWrite(audioLooper, nil); err != nil {
		return "", err
	}

	p, err := t.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return p.SID(), nil
}

func (t *LoadTester) PublishVideoTrack(name, resolution, codec string) (string, error) {
	if !t.IsRunning() {
		return "", nil
	}

	fmt.Println("publishing video track -", t.room.LocalParticipant.Identity())
	loopers, err := provider2.CreateVideoLoopers(resolution, codec, false)
	if err != nil {
		return "", err
	}
	track, err := lksdk.NewLocalSampleTrack(loopers[0].Codec())
	if err != nil {
		return "", err
	}
	if err := track.StartWrite(loopers[0], nil); err != nil {
		return "", err
	}

	p, err := t.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return p.SID(), nil
}

func (t *LoadTester) PublishData(packetSizeInByte, bitrate int, kind livekit.DataPacket_Kind, ready chan struct{}) error {
	if !t.IsRunning() {
		return nil
	}

	packetBits := packetSizeInByte * 8
	sendInterval := time.Duration(float64(time.Second) / float64(bitrate) * float64(packetBits))
	if sendInterval < time.Millisecond {
		return fmt.Errorf("packet size too small for bitrate, packets to send per second should be less than 1000")
	}

	data := prepareData(packetSizeInByte)

	fmt.Println("publishing data track -", t.room.LocalParticipant.Identity())
	if err := t.room.LocalParticipant.PublishData([]byte("ensure connect"), kind, []string{"unexist"}); err != nil {
		return err
	}
	go func() {
		if !t.dataPublishing.CompareAndSwap(false, true) {
			return // already publishing
		}
		ticker := time.NewTicker(sendInterval)
		defer func() {
			ticker.Stop()
			t.dataPublishing.Store(false)
		}()

		<-ready
		for range ticker.C {
			if !t.IsRunning() {
				return
			}

			data = prepareData(packetSizeInByte)

			err := t.room.LocalParticipant.PublishData(data, kind, nil)
			if err != nil {
				fmt.Println("error publishing data", err, "participant", t.room.LocalParticipant.Identity())
			}
		}
	}()
	return nil
}

func (t *LoadTester) PublishSimulcastTrack(name, resolution, codec string) (string, error) {
	var tracks []*lksdk.LocalSampleTrack

	fmt.Println("publishing simulcast video track -", t.room.LocalParticipant.Identity())
	loopers, err := provider2.CreateVideoLoopers(resolution, codec, true)
	if err != nil {
		return "", err
	}
	// for video, publish three simulcast layers
	for _, looper := range loopers {
		layer := looper.ToLayer()

		track, err := lksdk.NewLocalSampleTrack(looper.Codec(),
			lksdk.WithSimulcast("loadtest-video", layer))
		if err != nil {
			return "", err
		}
		if err := track.StartWrite(looper, nil); err != nil {
			return "", err
		}
		tracks = append(tracks, track)
	}

	p, err := t.room.LocalParticipant.PublishSimulcastTrack(tracks, &lksdk.TrackPublicationOptions{
		Name:   name,
		Source: livekit.TrackSource_CAMERA,
	})
	if err != nil {
		return "", err
	}

	return p.SID(), nil
}

func (t *LoadTester) getStats() *testerStats {
	stats := &testerStats{
		expectedTracks: t.params.expectedTracks,
		trackStats:     t.stats.Load().(*trackStats),
	}

	return stats
}

func (t *LoadTester) Reset() {
	stats := t.stats.Load().(*trackStats)

	var trackID string
	if stats != nil {
		trackID = stats.trackID
	}

	t.stats.Store(&trackStats{
		trackID: trackID,
	})
}

func (t *LoadTester) Stop() {
	if !t.IsRunning() {
		return
	}
	t.running.Store(false)
	t.room.Disconnect()
}

func (t *LoadTester) onTrackPublished(publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	publication.SetSubscribed(true)
}

func (t *LoadTester) onTrackSubscribed(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	value := t.stats.Load()
	stats, ok := value.(*trackStats)
	if !ok {
		stats = &trackStats{
			trackID: track.ID(),
		}

		t.stats.Store(stats)
	}

	stats.kind = pub.Kind()

	fmt.Println("subscribed to track", t.room.LocalParticipant.Identity(), pub.SID(), pub.Kind())

	go t.consumeTrack(track, pub, rp)

	if pub.Kind() != lksdk.TrackKindVideo {
		return
	}

	resolutions := provider2.GetVideoResolution(t.params.Resolution)
	if resolutions == nil || len(resolutions) != 3 {
		fmt.Printf("invalid resolution %s\n", t.params.Resolution)
		return
	}

	switch t.quality {
	case livekit.VideoQuality_HIGH:
		pub.SetVideoDimensions(uint32(resolutions[0].Width), uint32(resolutions[0].Height))
	case livekit.VideoQuality_MEDIUM:
		pub.SetVideoDimensions(uint32(resolutions[1].Width), uint32(resolutions[1].Height))
	case livekit.VideoQuality_LOW:
		pub.SetVideoDimensions(uint32(resolutions[2].Width), uint32(resolutions[2].Height))
	}
}

func (t *LoadTester) consumeTrack(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	rp.WritePLI(track.SSRC())

	defer func() {
		if e := recover(); e != nil {
			fmt.Println("caught panic in consumeTrack", e)
		}
	}()

	var dpkt rtp.Depacketizer
	isVideo := false
	if pub.Kind() == lksdk.TrackKindVideo {
		dpkt = &codecs.H264Packet{}
		isVideo = true
	} else {
		dpkt = &codecs.OpusPacket{}
	}

	value := t.stats.Load()
	stats, ok := value.(*trackStats)
	if !ok {
		fmt.Println("invalid stats")
		return
	}

	sb := samplebuilder.New(100, dpkt, track.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
		stats.dropped.Inc()
		if isVideo {
			rp.WritePLI(track.SSRC())
		}
	}))

	stats.startedAt.Store(time.Now())
	for {
		pkt, _, err := track.ReadRTP()
		if err != nil {
			return
		}
		if pkt == nil {
			continue
		}
		sb.Push(pkt)

		for _, pkt := range sb.PopPackets() {
			stats.bytes.Add(int64(len(pkt.Payload)))
			stats.packets.Inc()

			if pub.Kind() == lksdk.TrackKindVideo && len(pkt.Payload) > 8 {
				sentAt := int64(binary.LittleEndian.Uint64(pkt.Payload[len(pkt.Payload)-8:]))
				latency := time.Now().UnixNano() - sentAt
				sentTime := time.Unix(0, sentAt)

				// Define a reasonable time range for validation
				minTime := time.Now().Add(-20 * time.Minute)
				maxTime := time.Now().Add(20 * time.Minute)

				// Check if sentTime is within the valid range
				if sentTime.After(minTime) && sentTime.Before(maxTime) {
					if latency > 0 {
						stats.latency.Add(latency)
						stats.latencyCount.Inc()
					}
				}
			}
		}
	}
}

func (t *LoadTester) onDataReceived(data []byte, rp *lksdk.RemoteParticipant) {
	value := t.stats.Load()
	stats, ok := value.(*trackStats)
	if !ok {
		stats = &trackStats{
			trackID: rp.SID(),
		}

		t.stats.Store(stats)
	}

	load := t.dataChannelLoaded.Load()
	if !load {
		t.dataChannelLoaded.Store(true)
		stats.dataChannelStartedAt.Store(time.Now())
	}

	stats.dataChannelBytes.Add(int64(len(data)))
	stats.dataChannelPackets.Inc()
	if len(data) > 8 {
		// Extract the timestamp from the data
		sentAt := int64(binary.LittleEndian.Uint64(data[len(data)-8:]))

		// Calculate the latency
		latency := time.Now().UnixNano() - sentAt
		stats.dataChannelLatency.Add(latency)
		stats.dataChannelLatencyCount.Inc()
	}
}

func prepareData(size int) []byte {
	data := make([]byte, size)

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(time.Now().UnixNano()))

	return append(data, ts...)
}
