package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/abdooman21/go-discord/bot"
	"github.com/abdooman21/go-discord/bot/voice"
	"github.com/abdooman21/go-discord/internal/database"
	"github.com/abdooman21/go-discord/internal/env"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}

	ds, err := discordgo.New("Bot " + env.GetString("DISCORD_TOKEN", ""))

	if err != nil {
		log.Fatal(err)
	}

	dbUrl := os.Getenv("DB_URL")

	if dbUrl == "" {
		log.Fatal("Couldn't load database check .env")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	db := database.New(conn)

	app := bot.Application{
		Bot: ds,
		DB:  db,

		//added
		Voice: voice.NewVoiceManager(),
	}

	slog.Info("bot running !!")
	app.Run()

}
