package bot

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"

	"github.com/abdooman21/go-discord/quiz"
	"github.com/abdooman21/go-discord/web"

	"github.com/bwmarrin/discordgo"
)

var (
	sessionScores = make(map[string]int) // Key: UserID, Value: Score
	scoreMu       sync.Mutex
)

func (api *Application) HandleInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	// #TODO make the quiz handler in it's own function
	if !strings.HasPrefix(customID, "quiz_") {
		return
	}

	parts := strings.Split(customID, "_")
	choice, _ := strconv.Atoi(parts[1])
	correct, _ := strconv.Atoi(parts[2])

	user := i.Member.User
	if i.User != nil {
		user = i.User
	}

	if choice == correct {
		// err := db.AddPoint(context.Background(), database.AddPointParams{
		// 	UserID:   user.ID,
		// 	Username: user.Username,
		// })

		// if err != nil {
		// 	log.Println("Score Update Error:", err)
		// }
		scoreMu.Lock()
		sessionScores[user.ID]++
		currentScore := sessionScores[user.ID]
		scoreMu.Unlock()
		actionRow := i.Message.Components[0].(*discordgo.ActionsRow)
		correctBtn := actionRow.Components[correct].(*discordgo.Button)

		oldembed := i.Message.Embeds[0]
		oldembed.Description += fmt.Sprintf("\n الجواب الصحيح هو : %s ", correctBtn.Label)
		oldembed.Color = 0x00FF00 // Green

		msg := fmt.Sprintf(" إجابة صحيحة من **%s**! حصلت على نقطة.", user.Username)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Embeds: []*discordgo.MessageEmbed{{
					Title:       "🏆 بطل الكويز",
					Description: fmt.Sprintf("**%s** إجابته صحيحة! ونقاطك الحالية هي %d ", user.Username, currentScore),
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: user.AvatarURL("128"),
					},
					Color: 0xFFFF00, // Gold
				}},
			},
		})

		// disable buttons
		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         i.Message.ID,
			Channel:    i.ChannelID,
			Embeds:     &[]*discordgo.MessageEmbed{oldembed},
			Components: &[]discordgo.MessageComponent{},
		})

	} else {
		//  Send ephemeral message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: " إجابة خاطئة! حاول مرة أخرى.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
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

			}

		}

	}
}
