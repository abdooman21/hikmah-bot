package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (api *Application) slashRadio(s *discordgo.Session, i *discordgo.InteractionCreate) {
	voiceChannelID := findUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if voiceChannelID == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ يجب أن تكون في قناة صوتية أولاً.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	radios, err := api.DB.GetRadios(context.Background())
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ حدث خطأ أثناء جلب قائمة الإذاعات.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var options []discordgo.SelectMenuOption
	for _, r := range radios {
		options = append(options, discordgo.SelectMenuOption{
			Label:       r.Name,
			Value:       fmt.Sprintf("play_radio_%d", r.ID),
			Description: fmt.Sprintf("إذاعة رقم %d", r.ID),
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "📻 اختر الإذاعة التي تريد الاستماع إليها:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "radio_select",
							Placeholder: "اختر إذاعة...",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func (api *Application) slashStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, ok := api.Voice.GetPlayer(i.GuildID); !ok {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ البوت ليس في أي قناة صوتية.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	api.Voice.Leave(i.GuildID)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "⏹️ تم إيقاف البث وقطع الاتصال."},
	})
}

func (api *Application) slashHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	helpEmbed := &discordgo.MessageEmbed{
		Title: "📖 الأوامر المتاحة",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/سؤال", Value: "سؤال عشوائي واحد\nاختياري: `/سؤال [category] [difficulty]`", Inline: false},
			{Name: "/quiz", Value: "جلسة 5 جولات — ستظهر أزرار لاختيار التصنيف والصعوبة", Inline: false},
			{Name: "/راديو", Value: "تشغيل إذاعة إسلامية", Inline: false},
			{Name: "/وقف", Value: "إيقاف البث", Inline: false},
			{Name: "التصنيفات (1-6)", Value: "١ التفسير · ٢ العقيدة · ٣ الحديث · ٤ الفقه · ٥ التاريخ · ٦ اللغة_العربية", Inline: false},
			{Name: "الصعوبة (1-3)", Value: "١ سهل · ٢ متوسط · ٣ صعب", Inline: false},
		},
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{helpEmbed},
		},
	})
}
