package main

import (
	"log"

	"github.com/abdooman21/go-discord/bot"
	"github.com/abdooman21/go-discord/internal/env"
	"github.com/bwmarrin/discordgo"
)

func main() {

	ds, err := discordgo.New("Bot " + env.GetString("DISCORD_TOKEN", ""))

	if err != nil {
		log.Fatal(err)
	}
	app := bot.Application{
		Bot: ds,
	}

	log.Println("bot running !!")
	app.Run()

}
