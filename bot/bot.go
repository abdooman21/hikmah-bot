package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/abdooman21/go-discord/bot/voice"
	"github.com/abdooman21/go-discord/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Application struct {
	Bot *discordgo.Session
	DB  *database.Queries

	// added
	Voice *voice.VoiceManager
}

func (api *Application) Run() {

	api.Bot.AddHandler(api.newMessage)
	api.Bot.AddHandler(api.HandleInteractions)

	api.Bot.Open()
	defer api.Bot.Close()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	sig := <-c
	log.Println("Graceful  server kill", sig)

}
