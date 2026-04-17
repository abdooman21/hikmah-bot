package bot

import (
	"math/rand/v2"
	"strings"

	"github.com/abdooman21/go-discord/quiz"
	"github.com/abdooman21/go-discord/web"

	"github.com/bwmarrin/discordgo"
)

func (api *Application) HandleInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	// #TODO make the quiz handler in it's own function
	checkID := strings.Split(customID, "_")
	switch checkID[0] {
	case "quiz":
		quiz.QuizInteractionHandler(s, i, api.DB)
	case "session":
		quiz.SessionInteractionHandler(s, i, api.DB)
	case "setup":
		quiz.SetupInteractionHandler(s, i, api.DB)
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "oops where I am \"eyes\" ", Flags: 64},
		})
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

			case "Cat":
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

			case "مساعدة":
				helpEmbed := &discordgo.MessageEmbed{
					Title: "📖 كيف تستخدمني",
					Color: 0x5865F2,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "@bot سؤال", Value: "سؤال عشوائي واحد\nاختياري: `@bot سؤال [1-6] [1-3]` (تصنيف ثم صعوبة)", Inline: false},
						{Name: "@bot كويز", Value: "جلسة 5 جولات — ستظهر أزرار لاختيار التصنيف والصعوبة", Inline: false},
						{Name: "التصنيفات (1-6)", Value: "١ التفسير · ٢ العقيدة · ٣ الحديث · ٤ الفقه · ٥ التاريخ · ٦ اللغة_العربية", Inline: false},
						{Name: "الصعوبة (1-3)", Value: "١ سهل · ٢ متوسط · ٣ صعب", Inline: false},
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, helpEmbed)

				// added

			case "راديو":
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
						s.ChannelMessageSend(m.ChannelID, "❌ فشل الانضمام إلى القناة الصوتية.")
						return
					}
				}

				s.ChannelMessageSend(m.ChannelID, "📻 جارٍ تشغيل راديو القرآن الكريم...")

				player.PlayStream("https://Qurango.net/radio/saud_alshuraim")

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

				player.PlayStream("https://www.soundjay.com/buttons/sounds/beep-01a.mp3")

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
