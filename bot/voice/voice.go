package voice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cheezecakee/dca"
)

func defaultEncodeOptions() *dca.EncodeOptions {
	return &dca.EncodeOptions{
		FrameRate: 48000,

		FrameDuration: 20,

		Bitrate: 64,

		PacketLoss: 1,

		CompressionLevel: 10,

		BufferedFrames: 100,

		VariableBitrate: true,
	}
}

type VoiceManager struct {
	mu      sync.RWMutex
	players map[string]*Player
}

func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		players: make(map[string]*Player),
	}
}

type Player struct {
	mu        sync.Mutex
	guildID   string
	channelID string
	session   *discordgo.Session
	vc        *discordgo.VoiceConnection

	streamCtx    context.Context
	cancelStream context.CancelFunc
}

func (vm *VoiceManager) Join(s *discordgo.Session, guildID, channelID string) (*Player, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if p, ok := vm.players[guildID]; ok {
		p.Disconnect()
		delete(vm.players, guildID)
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("join failed: %w", err)
	}

	p := &Player{
		guildID:   guildID,
		channelID: channelID,
		session:   s,
		vc:        vc,
	}

	vm.players[guildID] = p
	return p, nil
}

func (vm *VoiceManager) GetPlayer(guildID string) (*Player, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	p, ok := vm.players[guildID]
	return p, ok
}

func (vm *VoiceManager) Leave(guildID string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if p, ok := vm.players[guildID]; ok {
		p.Disconnect()
		delete(vm.players, guildID)
	}
}

func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelStream != nil {
		p.cancelStream()
		p.cancelStream = nil
	}
}

func (p *Player) Disconnect() {
	p.Stop()
	p.vc.Speaking(false)
	p.vc.Disconnect()
}

func (p *Player) Play(url string) {
	p.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	p.mu.Lock()
	p.streamCtx = ctx
	p.cancelStream = cancel
	p.mu.Unlock()

	go p.loop(ctx, url)
}

func (p *Player) loop(ctx context.Context, url string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !p.vc.Ready {
			if err := p.reconnect(); err != nil {
				slog.Error("reconnect failed", "guild", p.guildID, "err", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}

		p.pipeline(ctx, url)

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (p *Player) reconnect() error {
	p.vc.Speaking(false)
	p.vc.Disconnect()

	vc, err := p.session.ChannelVoiceJoin(p.guildID, p.channelID, false, true)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.vc = vc
	p.mu.Unlock()

	return nil
}

func (p *Player) pipeline(ctx context.Context, url string) {
	deadline := time.Now().Add(10 * time.Second)
	for !p.vc.Ready {
		if time.Now().After(deadline) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
	}

	enc, err := dca.EncodeFile(url, defaultEncodeOptions())
	if err != nil {
		return
	}
	defer enc.Cleanup()

	if err := p.vc.Speaking(true); err != nil {
		return
	}
	defer p.vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	done := make(chan error, 1)
	dca.NewStream(enc, p.vc, done)

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			slog.Error("stream error", "guild", p.guildID, "err", err)
		}
	case <-ctx.Done():
		enc.Stop()
	}
}
