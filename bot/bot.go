package bot

import (
	"log/slog"
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
	// for dev
	// api.Bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
	// 	slog.Info("bot logged in", "username", r.User.Username)
	// 	guildID := os.Getenv("DEV_GUILD_ID")
	// 	api.RegisterCommands(guildID)
	// })

	err := api.Bot.Open()
	if err != nil {
		slog.Error("error opening connection", "err", err)
		return
	}
	defer api.Bot.Close()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	sig := <-c
	slog.Info("Graceful server kill", "signal", sig)

}
