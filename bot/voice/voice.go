package voice

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jogramming/dca"
)

const radioURL = "https://qurango.net/radio/fatwa"

type VoiceManager struct {
	mu          sync.Mutex
	connections map[string]*VoiceStream
}

type VoiceStream struct {
	vc   *discordgo.VoiceConnection
	stop chan struct{}
	done chan struct{}
}

func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		connections: make(map[string]*VoiceStream),
	}
}

func (vm *VoiceManager) Join(s *discordgo.Session, guildID, channelID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if existing, ok := vm.connections[guildID]; ok {
		vm.stopStream(existing)
		delete(vm.connections, guildID)
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	stream := &VoiceStream{
		vc:   vc,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	vm.connections[guildID] = stream
	go vm.streamAudio(stream)
	return nil
}

func (vm *VoiceManager) JoinTest(s *discordgo.Session, guildID, channelID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if existing, ok := vm.connections[guildID]; ok {
		vm.stopStream(existing)
		delete(vm.connections, guildID)
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	stream := &VoiceStream{
		vc:   vc,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	vm.connections[guildID] = stream
	go vm.PlayTestTone(stream)
	return nil
}

func (vm *VoiceManager) Leave(guildID string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	stream, ok := vm.connections[guildID]
	if !ok {
		return
	}
	vm.stopStream(stream)
	delete(vm.connections, guildID)
}

func (vm *VoiceManager) IsActive(guildID string) bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	_, ok := vm.connections[guildID]
	return ok
}

func (vm *VoiceManager) stopStream(stream *VoiceStream) {
	close(stream.stop)
	<-stream.done
	stream.vc.Speaking(false)
	stream.vc.Disconnect()
}

// ... [Join, JoinTest, Leave, IsActive, stopStream remain unchanged] ...

// runPipeline is the shared pipeline that extracts discrete Opus frames
// using the DCA package and sends them to Discord.
func (vm *VoiceManager) runPipeline(stream *VoiceStream, url string) {
	// Wait for voice connection ready — 10s timeout
	deadline := time.Now().Add(10 * time.Second)
	for !stream.vc.Ready {
		if time.Now().After(deadline) {
			log.Println("voice: timed out waiting for ready")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "audio"

	// dca.EncodeFile manages the FFmpeg subprocess and Opus encoding for you
	encodeSession, err := dca.EncodeFile(url, options)
	if err != nil {
		log.Println("voice: dca encode error:", err)
		return
	}
	defer encodeSession.Cleanup()

	if err := stream.vc.Speaking(true); err != nil {
		log.Println("voice: speaking error:", err)
		return
	}
	defer stream.vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)

	for {
		select {
		case <-stream.stop:
			return
		default:
		}

		// OpusFrame() reads EXACTLY one discrete frame.
		// It prevents slicing packets in half.
		frame, err := encodeSession.OpusFrame()
		if err != nil {
			if err != io.EOF {
				log.Println("voice: error reading opus frame:", err)
			}
			return
		}

		select {
		case stream.vc.OpusSend <- frame:
		case <-stream.stop:
			return
		}
	}
}

// streamAudio streams the live Quran radio station.
func (vm *VoiceManager) streamAudio(stream *VoiceStream) {
	defer close(stream.done)
	vm.runPipeline(stream, radioURL)
}

// PlayTestTone plays a test audio file to verify the pipeline.
func (vm *VoiceManager) PlayTestTone(stream *VoiceStream) {
	defer close(stream.done)

	// dca.EncodeFile expects an input URL or filepath. Since injecting raw "-f lavfi"
	// is tricky with the DCA wrapper, pointing it to a reliable remote MP3 or a local
	// test tone file is the cleanest way to test the pipeline.
	testToneURL := "https://actions.google.com/sounds/v1/alarms/beep_short.ogg"
	vm.runPipeline(stream, testToneURL)
}
