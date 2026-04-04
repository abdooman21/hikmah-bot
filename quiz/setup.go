package quiz

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var categories = map[int]string{
	1: "", 2: "", 3: "", 4: "", 5: "", 6: "",
}

func SendSetupMessage(s *discordgo.Session, channID string) {
	var catButtons []discordgo.MessageComponent
	for i := 1; i <= 6; i++ {
		catButtons = append(catButtons, discordgo.Button{
			Label:    fmt.Sprintf("%s", categories[i]),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("setup_cat_%d", i),
		})
	}

	diffButtons := []discordgo.MessageComponent{
		discordgo.Button{Label: "سهل", Style: discordgo.SuccessButton, CustomID: "setup_diff_1"},
		discordgo.Button{Label: "متوسط", Style: discordgo.PrimaryButton, CustomID: "setup_diff_2"},
		discordgo.Button{Label: "صعب", Style: discordgo.DangerButton, CustomID: "setup_diff_3"},
	}

	s.ChannelMessageSendComplex(channID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "🎮 إعداد الكويز",
			Description: "**١. اختر التصنيف:**",
			Color:       0x5865F2,
			Footer:      &discordgo.MessageEmbedFooter{Text: "اختر التصنيف ثم الصعوبة للبدء"},
		}},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: catButtons},
			discordgo.ActionsRow{Components: diffButtons},
		},
	})
}
