package voice

import (
	"context"
	"fmt"
	"io"
	"log/slog" // Good for professional logging
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cheezecakee/dca"
)

// defaultEncodeOptions holds the FFmpeg settings.
func defaultEncodeOptions() *dca.EncodeOptions {
	return &dca.EncodeOptions{

		FrameRate: 48000,

		FrameDuration: 20,

		Bitrate: 64,

		PacketLoss: 1,

		CompressionLevel: 10,

		BufferedFrames: 100,

		VariableBitrate: true,

		StartTime: 0,
	}
}

// VoiceManager keeps track of all the servers the bot is in.
type VoiceManager struct {
	mu      sync.RWMutex
	players map[string]*Player
}

func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		players: make(map[string]*Player),
	}
}

// Player controls the audio for one specific server.
type Player struct {
	mu      sync.Mutex
	guildID string
	vc      *discordgo.VoiceConnection

	// Context helps us stop the audio safely
	streamCtx    context.Context
	cancelStream context.CancelFunc
}

// Join makes the bot enter the voice channel.
func (vm *VoiceManager) Join(s *discordgo.Session, guildID, channelID string) (*Player, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// If already in this server, clean up first
	if existingPlayer, ok := vm.players[guildID]; ok {
		existingPlayer.Disconnect()
		delete(vm.players, guildID)
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to join: %w", err)
	}

	player := &Player{
		guildID: guildID,
		vc:      vc,
	}
	vm.players[guildID] = player

	return player, nil
}

// GetPlayer finds the bot in a server without joining again.
func (vm *VoiceManager) GetPlayer(guildID string) (*Player, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	p, ok := vm.players[guildID]
	return p, ok
}

// Leave stops everything and removes the bot from the channel.
func (vm *VoiceManager) Leave(guildID string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if player, ok := vm.players[guildID]; ok {
		player.Disconnect()
		delete(vm.players, guildID)
	}
}

// StopCurrentAudio stops the sound but keeps the bot in the channel.
func (p *Player) StopCurrentAudio() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelStream != nil {
		p.cancelStream()
		p.cancelStream = nil
	}
}

// Disconnect stops sound and leaves the channel.
func (p *Player) Disconnect() {
	p.StopCurrentAudio()
	p.vc.Speaking(false)
	p.vc.Disconnect()
}

// PlayStream plays any link you give it.
// If something is already playing, it stops the old one and starts the new one.
func (p *Player) PlayStream(audioURL string) {
	p.StopCurrentAudio()

	ctx, cancel := context.WithCancel(context.Background())

	p.mu.Lock()
	p.streamCtx = ctx
	p.cancelStream = cancel
	p.mu.Unlock()

	go p.runPipeline(ctx, audioURL)
}

func (p *Player) runPipeline(ctx context.Context, inputURL string) {
	deadline := time.Now().Add(10 * time.Second)
	for !p.vc.Ready {
		if time.Now().After(deadline) {
			slog.Error("timeout waiting for ready", "guild", p.guildID)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	encodeSession, err := dca.EncodeFile(inputURL, defaultEncodeOptions())
	if err != nil {
		slog.Error("dca error", "error", err)
		return
	}
	defer encodeSession.Cleanup()

	if err := p.vc.Speaking(true); err != nil {
		slog.Error("speaking error", "error", err)
		return
	}
	defer p.vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	done := make(chan error)
	dca.NewStream(encodeSession, p.vc, done)

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			slog.Error("stream error", "error", err)
		}
	case <-ctx.Done():
		// Stop was called
		encodeSession.Stop()
	}
}
