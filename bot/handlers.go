package bot

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/abdooman21/go-discord/quiz"
	"github.com/abdooman21/go-discord/web"

	"github.com/bwmarrin/discordgo"
)

func (api *Application) HandleInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommand:
		api.handleSlashCommand(s, i)

	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		checkID := strings.Split(customID, "_")
		switch checkID[0] {
		case "quiz":
			quiz.QuizInteractionHandler(s, i, api.DB)
		case "session":
			quiz.SessionInteractionHandler(s, i, api.DB)
		case "setup":
			quiz.SetupInteractionHandler(s, i, api.DB)
		case "radio":
			voiceHand(s, i, api, customID)
		default:
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "oops where I am \"eyes\" ", Flags: 64},
			})
		}
	}
}

func (api *Application) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "quiz":
		api.slashQuiz(s, i)
	case "سؤال":
		api.slashSingleQuestion(s, i)
	case "راديو":
		api.slashRadio(s, i)
	case "وقف":
		api.slashStop(s, i)
	case "مساعدة":
		api.slashHelp(s, i)
	}
}

func (api *Application) newMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		// for _, user := range m.Mentions {
		// 	if user.ID == (s.State.User.ID) {
		// 		s.ChannelMessageSend(m.ChannelID, "beep boop")
		// 	}
		// }
		// log.Println("bot talked: ", m.Author.Username)
		return
	}
	for _, user := range m.Mentions {
		if user.ID == (s.State.User.ID) {
			Param := strings.Fields(m.Content)

			if len(Param) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Nothing to see")
				return
			}

			switch Param[1] {

			case "hey":
				s.ChannelTyping(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, "hey")

			case "cat", "Cat", "قطة":
				fact := web.GetCatFact()
				if fact == "" {
					s.ChannelTyping(m.ChannelID)
					s.ChannelMessageSend(m.ChannelID, "Sorry an error occured")
					return
				}
				s.ChannelTyping(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fact)
			case "flip":
				s.ChannelTyping(m.ChannelID)
				luck := rand.IntN(2)

				if luck == 0 {
					s.ChannelMessageSend(m.ChannelID, "tail")
					return
				}
				s.ChannelMessageSend(m.ChannelID, "head")
			case "السلام":
				s.ChannelTyping(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, "وعليكم السلام ورحمة الله وبركاته")

				// case "weather", "الطقس":
				// 	s.ChannelTyping(m.ChannelID)
				// 	if len(Param) > 2 {
				// 		s.ChannelMessageSendComplex(m.ChannelID, web.GetCurrentWeather(Param[2]))
				// 		return
				// 	}

				// 	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				// 		Content: fmt.Sprintf("إYou should provide zip code after the, \"%v\" ", Param[1]),
				// 	})

			case "سؤال":
				go quiz.Get_Q(s, m, api.DB)

			case "كويز":
				go quiz.Start_session(s, m, api.DB)

			case "مساعدة", "مساعده", "help", "الو":
				helpEmbed := &discordgo.MessageEmbed{
					Title: "📖 كيف تستخدمني",
					Color: 0x5865F2,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "@bot سؤال", Value: "سؤال عشوائي واحد\nاختياري: `@bot سؤال [1-6] [1-3]` (تصنيف ثم صعوبة)", Inline: false},
						{Name: "@bot كويز", Value: "جلسة 5 جولات — ستظهر أزرار لاختيار التصنيف والصعوبة", Inline: false},
						{Name: "التصنيفات (1-6)", Value: "١ التفسير · ٢ العقيدة · ٣ الحديث · ٤ الفقه · ٥ التاريخ · ٦ اللغة_العربية", Inline: false},
						{Name: "الصعوبة (1-3)", Value: "١ سهل · ٢ متوسط · ٣ صعب", Inline: false},
						{Name: "قطة || cat", Value: "اجلب معلومة متعلقة عن القطط", Inline: false},
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, helpEmbed)

				// added

			case "راديو":
				radios, err := api.DB.GetRadios(context.Background())
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "❌ حدث خطأ أثناء جلب قائمة الإذاعات.")
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

				s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
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
				})

			case "تست":
				voiceChannelID := findUserVoiceChannel(s, m.GuildID, m.Author.ID)
				if voiceChannelID == "" {
					s.ChannelMessageSend(m.ChannelID, "⚠️ يجب أن تكون في قناة صوتية أولاً.")
					return
				}

				player, ok := api.Voice.GetPlayer(m.GuildID)
				if !ok {
					var err error
					player, err = api.Voice.Join(s, m.GuildID, voiceChannelID)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "❌ فشل الانضمام.")
						return
					}
				}

				s.ChannelMessageSend(m.ChannelID, "🔊 تشغيل نغمة اختبار...")

				player.Play("https://www.soundjay.com/buttons/sounds/beep-01a.mp3")

			case "وقف":
				api.Voice.Leave(m.GuildID)
				s.ChannelMessageSend(m.ChannelID, "⏹️ تم إيقاف البث وقطع الاتصال.")

			}

		}

	}
}

// findUserVoiceChannel returns the voice channel ID the user is currently in,
// or an empty string
func findUserVoiceChannel(s *discordgo.Session, guildID, userID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return ""
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID
		}
	}
	return ""
}

func voiceHand(s *discordgo.Session, i *discordgo.InteractionCreate, api *Application, customID string) {
	if customID == "radio_select" {
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			return
		}
		val := values[0]
		parts := strings.Split(val, "_")
		if len(parts) < 3 {
			return
		}
		radioID, err := strconv.Atoi(parts[2])
		if err != nil {
			return
		}

		radio, err := api.DB.GetRadioByID(context.Background(), int32(radioID))
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "❌ لم يتم العثور على الإذاعة."},
			})
			return
		}

		voiceChannelID := findUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
		if voiceChannelID == "" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "⚠️ يجب أن تكون في قناة صوتية أولاً."},
			})
			return
		}

		player, ok := api.Voice.GetPlayer(i.GuildID)
		if !ok {
			player, err = api.Voice.Join(s, i.GuildID, voiceChannelID)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: "❌ فشل الانضمام إلى القناة الصوتية."},
				})
				return
			}
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    fmt.Sprintf("📻 جارٍ تشغيل: **%s**", radio.Name),
				Components: []discordgo.MessageComponent{},
			},
		})

		player.Play(radio.Link)
	}
}
