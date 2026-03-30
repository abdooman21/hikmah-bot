package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/abdooman21/go-discord/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Answer struct {
	Text      string `json:"text"`
	IsCorrect int    `json:"is_correct"`
}

func Start_session(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries) {

	args := strings.Fields(m.Content)
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "الأستخدام \" كويز <رقم الفئة> <مستوى السؤال من 3>\"")
		return
	}

	catID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Print("failed at catid, ", catID)
		s.ChannelMessageSend(m.ChannelID, "!4## فشل بجلب السؤال خطا من المدخلات")
		return
	}
	level, err := strconv.Atoi(args[2])
	if err != nil {
		log.Print("failed at catid, ", catID)
		s.ChannelMessageSend(m.ChannelID, "!4## فشل بجلب السؤال خطا من المدخلات")
		return

	}
	ctx := context.Background()
	qData, err := db.GetRandomQByCatnLvl(ctx, database.GetRandomQByCatnLvlParams{
		ID:          int32(catID),
		LevelNumber: int32(level),
	})

	if err != nil {
		log.Println("failed at getting Q, ", err.Error(), "params : cat Id ", catID, " level: ", level)
		s.ChannelMessageSend(m.ChannelID, "!5## فشل بجلب السؤال المعذرة على الخطأ ")
		return
	}

	var answers []Answer
	err = json.Unmarshal(qData.Answers, &answers)
	if err != nil {
		log.Println("JSON Error:", err)
		return
	}

	// format
	answerText := ""
	for i, a := range answers {
		answerText += fmt.Sprintf("**%d.** %s\n", i+1, a.Text)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Topic: %s", qData.TopicName),
		Description: fmt.Sprintf("### %s\n\n%s", qData.QText, answerText),
		Color:       0x00ff00, // Green
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: qData.IconPath.String,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "React with the correct number or click a button!",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
