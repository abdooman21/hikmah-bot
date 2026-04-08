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
	// parts: ["setup", "go|cat"|"diff", value]
	channID := i.ChannelID

	user := i.Member.User
	if i.User != nil {
		user = i.User
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

	if parts[1] == "go" {
		if len(p.CategoryID) == 0 {
			pendingMu.Unlock()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ يرجى اختيار تصنيف واحد على الأقل!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		} else if p.Difficulty == 0 {
			pendingMu.Unlock()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ يرجى تحديد صعوبة", // Remove this later
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		cat, diff := p.CategoryID, p.Difficulty
		delete(pendingSetups, channID)
		pendingMu.Unlock()

		// 1. EDIT the original setup message to say "Started"
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{{
					Title:       "🎮 بدأ الكويز!",
					Description: "تم إغلاق الإعدادات، انطلقت المسابقة الآن...",
					Color:       0x00FF00,
				}},
				Components: []discordgo.MessageComponent{},
			},
		})

		startSessionWithParams(s, channID, user.ID, db, cat, diff)
		return
	}

	if parts[1] == "cat" {
		v, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Println("failed at cat/btn conversion for , ", parts[2])
			return
		}
		found := false
		if v == 9 {

			if len(p.CategoryID) == 6 {
				p.CategoryID = []int{}
			} else {
				p.CategoryID = []int{1, 2, 3, 4, 5, 6}
			}
			goto C1
		} else {
			for idx, existing := range p.CategoryID {
				if existing == v {
					p.CategoryID = append(p.CategoryID[:idx], p.CategoryID[idx+1:]...)
					found = true
					break
				}
			}
		}
		if !found {
			p.CategoryID = append(p.CategoryID, v)
		}
	}

	if parts[1] == "diff" {
		v, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Println("failed at diff/btn conversion for , ", parts[2])
			return
		}
		p.Difficulty = v
	}
C1:
	selectedCats := ""
	for _, id := range p.CategoryID {
		selectedCats += categories[id] + " ، "
	}
	if selectedCats == "" {
		selectedCats = "لم يتم الاختيار"
	}

	diffText := []string{"غير محدد", "سهل", "متوسط", "صعب"}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title: "🎮 إعداد الكويز",
				Description: fmt.Sprintf("**التصنيفات المختارة:** %s\n**الصعوبة:** %s",
					selectedCats, diffText[p.Difficulty]),
				Color: 0x5865F2,
			}},
			Components: i.Message.Components,
		},
	})
	pendingMu.Unlock()
}
