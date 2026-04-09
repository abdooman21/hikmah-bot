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
	allbtn := discordgo.Button{
		Label:    "الكل",
		Style:    discordgo.PrimaryButton,
		CustomID: fmt.Sprintf("setup_cat_%d", 9),
	}
	row2 = append(row2, allbtn)

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
	action, valStr := parts[1], ""
	if len(parts) > 2 {
		valStr = parts[2]
	}

	pendingMu.Lock()
	defer pendingMu.Unlock()

	p, ok := pendingSetups[i.ChannelID]
	if !ok {
		p = &pendingSetup{StartedBy: i.Member.User.ID}
		pendingSetups[i.ChannelID] = p
	}

	if p.StartedBy != i.Member.User.ID && i.User != nil && p.StartedBy != i.User.ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⛔ التعديل متاح فقط لمن بدأ الجلسة.", Flags: 64},
		})
		return
	}

	switch action {
	case "go":
		if len(p.CategoryID) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "⚠️ اختر تصنيفاً واحداً على الأقل!", Flags: 64},
			})
			return
		}
		delete(pendingSetups, i.ChannelID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{{Title: "🎮 بدأ الكويز!", Color: 0x00FF00}},
				Components: []discordgo.MessageComponent{},
			},
		})
		go startSessionWithParams(s, i.ChannelID, p.StartedBy, db, p.CategoryID, p.Difficulty)
		return

	case "cat":
		v, _ := strconv.Atoi(valStr)
		if v == 9 { // select ALL
			if len(p.CategoryID) == 6 {
				p.CategoryID = []int{}
			} else {
				p.CategoryID = []int{1, 2, 3, 4, 5, 6}
			}
		} else {
			found := false
			for idx, id := range p.CategoryID {
				if id == v {
					p.CategoryID = append(p.CategoryID[:idx], p.CategoryID[idx+1:]...)
					found = true
					break
				}
			}
			if !found {
				p.CategoryID = append(p.CategoryID, v)
			}
		}

	case "diff":
		v, _ := strconv.Atoi(valStr)
		p.Difficulty = v
	}

	selected := ""
	for _, id := range p.CategoryID {
		selected += categories[id] + " ، "
	}
	if selected == "" {
		selected = "لم يتم الاختيار"
	}
	diffs := []string{"عشوائي 🎲", "سهل 🟢", "متوسط 🟡", "صعب 🔴"}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "🎮 إعداد الكويز",
				Description: fmt.Sprintf("**التصنيفات:** %s\n**الصعوبة:** %s", selected, diffs[p.Difficulty]),
				Color:       0x5865F2,
			}},
			Components: i.Message.Components,
		},
	})
}
