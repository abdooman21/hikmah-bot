package bot

import (
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"strings"

	"github.com/abdooman21/go-discord/quiz"
	"github.com/abdooman21/go-discord/web"
	"github.com/bwmarrin/discordgo"
)

type Application struct {
	Bot *discordgo.Session
}

func (api *Application) Run() {

	api.Bot.AddHandler(newMessage)

	api.Bot.Open()
	defer api.Bot.Close()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	sig := <-c
	log.Println("Graceful  server kill", sig)

}

func newMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			case "كويز":
				go quiz.Start_session(s, m)

			}

		}

	}
}
