package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/abdooman21/go-discord/bot"
	"github.com/abdooman21/go-discord/internal/database"
	"github.com/abdooman21/go-discord/internal/env"
	"github.com/bwmarrin/discordgo"
)

func main() {

	ds, err := discordgo.New("Bot " + env.GetString("DISCORD_TOKEN", ""))

	if err != nil {
		log.Fatal(err)
	}
	dbUrl := os.Getenv("DB_URL")

	if dbUrl == "" {
		log.Fatal("Couldn't load database")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	db := database.New(conn)

	app := bot.Application{
		Bot: ds,
		DB:  db,
	}

	log.Println("bot running !!")
	app.Run()

}
