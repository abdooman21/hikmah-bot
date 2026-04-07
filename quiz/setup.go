package quiz

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/abdooman21/go-discord/internal/database"
	"github.com/bwmarrin/discordgo"
)

var categories = map[int]string{
	1: "التفسير", 2: "العقيدة", 3: "الحديث", 4: "الفقه", 5: "التاريخ", 6: "اللغة العربية",
}

func SendSetupMessage(s *discordgo.Session, channID string) {
	var row1 []discordgo.MessageComponent
	var row2 []discordgo.MessageComponent

	for i := 1; i <= 6; i++ {
		btn := discordgo.Button{
			Label:    categories[i],
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("setup_cat_%d", i),
		}

		if i <= 5 {
			row1 = append(row1, btn)
		} else {
			row2 = append(row2, btn)
		}
	}

	diffButtons := []discordgo.MessageComponent{
		discordgo.Button{Label: "سهل", Style: discordgo.SuccessButton, CustomID: "setup_diff_1"},
		discordgo.Button{Label: "متوسط", Style: discordgo.PrimaryButton, CustomID: "setup_diff_2"},
		discordgo.Button{Label: "صعب", Style: discordgo.DangerButton, CustomID: "setup_diff_3"},
	}
	gobtn := discordgo.Button{Label: "أنطلق!", Style: discordgo.DangerButton, CustomID: "setup_go"}

	_, err := s.ChannelMessageSendComplex(channID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "🎮 إعداد الكويز",
			Description: "**١. اختر التصنيف:**",
			Color:       0x5865F2,
			Footer:      &discordgo.MessageEmbedFooter{Text: "اختر التصنيف ثم الصعوبة للبدء"},
		}},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: row1},
			discordgo.ActionsRow{Components: row2},
			discordgo.ActionsRow{Components: diffButtons},
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{gobtn}},
		},
	})

	if err != nil {
		log.Println("SendSetupMessage failed:", err)
	}
}

type pendingSetup struct {
	CategoryID []int
	Difficulty int
	StartedBy  string
}

var pendingSetups = make(map[string]*pendingSetup) // channelID -> setup
var pendingMu sync.Mutex

func SetupInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Queries) {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	// parts: ["setup", "cat"|"diff", value]
	channID := i.ChannelID

	user := i.User
	if i.Member != nil {
		user = i.Member.User
	}
	if user == nil {
		return
	}

	pendingMu.Lock()
	p, ok := pendingSetups[channID]
	if !ok {
		p = &pendingSetup{StartedBy: user.ID}
		pendingSetups[channID] = p
	}

	if p.StartedBy != user.ID {
		pendingMu.Unlock()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⛔ فقط من بدأ الكويز يمكنه اختيار الإعدادات.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch parts[1] {
	case "cat":
		v, _ := strconv.Atoi(parts[2])
		p.CategoryID = append(p.CategoryID, v)
	case "diff":
		v, _ := strconv.Atoi(parts[2])
		p.Difficulty = v
	}

	ready := p.Difficulty > 0
	cat, diff := p.CategoryID, p.Difficulty
	if ready {
		delete(pendingSetups, channID)
	}
	pendingMu.Unlock()

	if !ready {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "✅ اختيارك سُجِّل، أكمل الإعداد.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Both selections done — acknowledge and start
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	startSessionWithParams(s, channID, user.ID, db, cat, diff)
}
