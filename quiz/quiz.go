package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"github.com/abdooman21/go-discord/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Answer struct {
	Text string `json:"answer"`
	True int    `json:"t"`
}

func send_QwithCriteria(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries, cat, lvl int) {

	ctx := context.Background()
	qData, err := db.GetRandomQByCatnLvl(ctx, database.GetRandomQByCatnLvlParams{
		ID:          int32(cat),
		LevelNumber: int32(lvl),
	})

	if err != nil {
		log.Println("failed at getting Q, ", err.Error(), "params : cat Id ", cat, " level: ", lvl)
		s.ChannelMessageSend(m.ChannelID, "!5## فشل بجلب السؤال المعذرة على الخطأ ")
		return
	}

	var answers []Answer
	err = json.Unmarshal(qData.Answers, &answers)
	if err != nil {
		log.Println("JSON Error:", err)
		return
	}

	var buttons []discordgo.MessageComponent
	var correctIndex int

	// answerText := ""
	for i, a := range answers {

		if a.True == 1 {
			correctIndex = i // another loop to setup index
		}
	}

	for i, a := range answers {

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%d. %s", i+1, a.Text),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("quiz_%d_%d", i, correctIndex),
		})
	}

	temppath := "internal/database/icons/asteroid.png"
	name := "icon"
	file, err := os.Open(temppath)
	if err != nil {
		log.Print("failed to retreve path : ", temppath)
		return
	}

	defer file.Close()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Topic: %s", qData.TopicName),
		Description: fmt.Sprintf("### %s", qData.QText),
		Color:       0x00ff00, // Green
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "attachment://" + name,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "React with the correct number or click a button!",
		},
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:   name,
				Reader: file,
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: buttons},
		},
	})

}

func Get_Q(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries) {
	args := strings.Fields(m.Content)

	// args[0] = @bot
	// args[1] = سؤال
	// args[2] = catID
	// args[3] = level

	if len(args) < 4 {
		cat := rand.IntN(6) + 1
		lvl := rand.IntN(3) + 1
		send_QwithCriteria(s, m, db, cat, lvl)
		s.ChannelMessageSend(m.ChannelID, " للاختيار: `@البوت سؤال <الفئة 1-6> <المستوى 1-3>` 4##")
		return
	}

	catID, err1 := strconv.Atoi(args[2])
	level, err2 := strconv.Atoi(args[3])

	if err1 != nil || err2 != nil || catID > 7 || catID < 1 || level > 3 || level < 1 {
		s.ChannelMessageSend(m.ChannelID, " خطأ في المدخلات: تأكد من كتابة أرقام صحيحة للفئة والمستوى 4##")
		return
	}

	send_QwithCriteria(s, m, db, catID, level)
}

func Start_session(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries) {
	args := strings.Fields(m.Content)

	// args[0] = @bot
	// args[1] = كويز
	// args[2] = catID
	// args[3] = level

	if len(args) < 4 {
		s.ChannelMessageSend(m.ChannelID, " الأستخدام: `@البوت كويز <رقم الفئة> <المستوى>` 4##")
		return
	}

	catID, err1 := strconv.Atoi(args[2])
	level, err2 := strconv.Atoi(args[3])

	if err1 != nil || err2 != nil || catID > 7 || catID < 1 || level > 3 || level < 1 {
		s.ChannelMessageSend(m.ChannelID, "!4## خطا بالمدخلات")
		return
	}

	send_QwithCriteria(s, m, db, catID, level)
}
