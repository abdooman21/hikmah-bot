package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abdooman21/go-discord/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Answer struct {
	Text string `json:"answer"`
	True int    `json:"t"`
}

type QuizSession struct {
	ChannelID string
	MessageID string // question message
	StartedBy string

	CurrentRound int
	MaxRounds    int

	CategoryID []int
	Difficulty int

	Scores       map[string]int    // UserID -> Score
	Participants map[string]string // UserID -> Username

	IsActive bool
	Mutex    sync.Mutex
}

var QScores = make(map[string]int)
var Qmu sync.RWMutex

// Global map to track sessions per channel
var activeSessions = make(map[string]*QuizSession)
var sessionsMu sync.RWMutex

func send_QwithCriteria(s *discordgo.Session, channID string, db *database.Queries, session string, cat []int, lvl int) {
	var catID int
	if len(cat) <= 0 {
		catID = rand.IntN(6) + 1
	} else {
		catID = cat[rand.IntN(len(cat))]
	}

	targetLvl := lvl
	if lvl == 0 {
		targetLvl = rand.IntN(3) + 1
	}

	ctx := context.Background()
	qData, err := db.GetRandomQByCatnLvl(ctx, database.GetRandomQByCatnLvlParams{
		ID:          int32(catID),
		LevelNumber: int32(targetLvl),
	})

	if err != nil {
		slog.Error("failed at getting Q", "err", err, "cat", cat, "level", lvl)
		s.ChannelMessageSend(channID, "!5## فشل بجلب السؤال المعذرة على الخطأ ")

		return
	}

	var answers []Answer
	err = json.Unmarshal(qData.Answers, &answers)
	if err != nil {
		slog.Error("JSON Error", "err", err)
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
			CustomID: fmt.Sprintf("%s_%d_%d", session, i, correctIndex),
		})
	}

	temppath := "internal/database/icons/asteroid.png"
	name := "icon"
	file, err := os.Open(temppath)
	if err != nil {
		slog.Error("failed to retrieve path", "path", temppath)
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

	_, err = s.ChannelMessageSendComplex(channID, &discordgo.MessageSend{
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
	if err != nil {
		slog.Error("failed to send message", "err", err)
	}

}

func Get_Q(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries) {
	/*
		args[0] = @bot
		args[1] = سؤال
		args[2] = catID
		args[3] = level
	*/
	sessionsMu.RLock()
	session, exists := activeSessions[m.ChannelID]
	sessionsMu.RUnlock()

	if exists && session.IsActive {
		msg := fmt.Sprintf(
			"<@%s>  لا يمكن توليد سؤال أثناء وجود جلسة كويز فعالة.",
			m.Author.ID,
		)

		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}
	args := strings.Fields(m.Content)

	cat := rand.IntN(6) + 1
	lvl := rand.IntN(3) + 1

	if len(args) >= 3 {
		if v, err := strconv.Atoi(args[2]); err == nil && v >= 1 && v <= 6 {
			cat = v
		}
	}

	if len(args) >= 4 {
		if v, err := strconv.Atoi(args[3]); err == nil && v >= 1 && v <= 3 {
			lvl = v
		}
	}

	send_QwithCriteria(s, m.ChannelID, db, "quiz", []int{cat}, lvl)
}

func QuizInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Queries) {
	customID := i.MessageComponentData().CustomID
	parts := strings.Split(customID, "_")
	choice, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	correct, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	user := i.Member.User
	if i.User != nil {
		user = i.User
	}

	sessionsMu.RLock()
	session, exists := activeSessions[i.ChannelID]
	sessionsMu.RUnlock()

	if exists && session.IsActive {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "لا يمكن توليد اسئلة أثناء وجود جلسة اخرى فعالة", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}
	if choice == correct {

		actionRow := i.Message.Components[0].(*discordgo.ActionsRow)
		correctBtn := actionRow.Components[correct].(*discordgo.Button)

		oldembed := i.Message.Embeds[0]
		oldembed.Description += fmt.Sprintf("\n الجواب الصحيح هو : %s ", correctBtn.Label)
		oldembed.Color = 0x00FF00 // Green

		msg := fmt.Sprintf("✅ إجابة صحيحة من **%s**! ", user.Username)
		Qmu.Lock()
		QScores[user.Username] += 1
		scores := QScores[user.Username]
		Qmu.Unlock()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Embeds: []*discordgo.MessageEmbed{{
					Title:       "🏆 بطل الكويز",
					Description: fmt.Sprintf("**%s** إجابته صحيحة! ونقاطك الحالية هي **%d**", user.Username, scores), // # TODO add points
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: user.AvatarURL("128"),
					},
					Color: 0xFFFF00, // Gold
				}},
			},
		})

		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         i.Message.ID,
			Channel:    i.ChannelID,
			Embeds:     &[]*discordgo.MessageEmbed{oldembed},
			Components: &[]discordgo.MessageComponent{},
		})

	} else {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ خطأ! حاول مرة أخرى.", Flags: discordgo.MessageFlagsEphemeral},
		})
	}

}
func SessionInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Queries) {
	sessionsMu.RLock()
	session, exists := activeSessions[i.ChannelID]
	sessionsMu.RUnlock()

	if !exists || !session.IsActive {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "هذا الكويز انتهى بالفعل!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	data := i.MessageComponentData()
	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 3 {
		return
	}

	choice, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	correct, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	user := i.Member.User
	if i.User != nil {
		user = i.User
	}

	if choice != correct {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ خطأ! حاول مرة أخرى.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	session.Mutex.Lock()
	session.Scores[user.ID]++
	session.Participants[user.ID] = user.Username
	session.CurrentRound++

	curScore := session.Scores[user.ID]
	roundReached := session.CurrentRound
	maxRounds := session.MaxRounds
	cat := session.CategoryID
	lvl := session.Difficulty
	session.Mutex.Unlock()

	oldEmbed := i.Message.Embeds[0]

	var correctLabel string
	if len(i.Message.Components) > 0 {
		if row, ok := i.Message.Components[0].(*discordgo.ActionsRow); ok {
			if btn, ok := row.Components[correct].(*discordgo.Button); ok {
				correctLabel = btn.Label
			}
		}
	}

	oldEmbed.Description += fmt.Sprintf("\n✅ الجواب الصحيح هو: **%s**", correctLabel)
	oldEmbed.Color = 0x00FF00

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ إجابة صحيحة من **%s**! (جولة %d/%d)", user.Username, roundReached, maxRounds),
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "🏆 بطل الكويز",
				Description: fmt.Sprintf("**%s** حصل على نقطة! رصيدك: `%d`", user.Username, curScore),
				Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: user.AvatarURL("128")},
				Color:       0xFFFF00,
			}},
		},
	})

	s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         i.Message.ID,
		Channel:    i.ChannelID,
		Embeds:     &[]*discordgo.MessageEmbed{oldEmbed},
		Components: &[]discordgo.MessageComponent{},
	})

	if roundReached < maxRounds {
		go func() {
			time.Sleep(2 * time.Second)
			send_QwithCriteria(s, i.ChannelID, db, "session", cat, lvl)
		}()
	} else {
		finishSession(s, i.ChannelID, session)
	}
}

// func SessionInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Queries) {
// 	customID := i.MessageComponentData().CustomID
// 	parts := strings.Split(customID, "_")
// 	choice, _ := strconv.Atoi(parts[1])
// 	correct, _ := strconv.Atoi(parts[2])

// 	user := i.Member.User
// 	if i.User != nil {
// 		user = i.User
// 	}

// 	//  Get the session for this channel
// 	sessionsMu.RLock()
// 	session, exists := activeSessions[i.ChannelID]
// 	sessionsMu.RUnlock()

// 	if !exists || !session.IsActive {
// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{Content: "هذا الكويز انتهى بالفعل!", Flags: discordgo.MessageFlagsEphemeral},
// 		})
// 		return
// 	}

// 	if choice == correct {
// 		session.Mutex.Lock()
// 		session.Scores[user.ID]++
// 		curScore := session.Scores[user.ID]
// 		session.Participants[user.ID] = user.Username
// 		session.CurrentRound++
// 		roundReached := session.CurrentRound
// 		maxRounds := session.MaxRounds
// 		cat := session.CategoryID
// 		lvl := session.Difficulty
// 		session.Mutex.Unlock()

// 		actionRow := i.Message.Components[0].(*discordgo.ActionsRow)
// 		correctBtn := actionRow.Components[correct].(*discordgo.Button)

// 		oldembed := i.Message.Embeds[0]
// 		oldembed.Description += fmt.Sprintf("\n الجواب الصحيح هو : %s ", correctBtn.Label)
// 		oldembed.Color = 0x00FF00 // Green

// 		msg := fmt.Sprintf("✅ إجابة صحيحة من **%s**! (جولة %d/%d)", user.Username, roundReached, maxRounds)

// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{
// 				Content: msg,
// 				Embeds: []*discordgo.MessageEmbed{{
// 					Title:       "🏆 بطل الكويز",
// 					Description: fmt.Sprintf("**%s** إجابته صحيحة! ونقاطك الحالية هي %d ", user.Username, curScore),
// 					Thumbnail: &discordgo.MessageEmbedThumbnail{
// 						URL: user.AvatarURL("128"),
// 					},
// 					Color: 0xFFFF00, // Gold
// 				}},
// 			},
// 		})

// 		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
// 			ID:         i.Message.ID,
// 			Channel:    i.ChannelID,
// 			Embeds:     &[]*discordgo.MessageEmbed{oldembed},
// 			Components: &[]discordgo.MessageComponent{},
// 		})

// 		if roundReached < maxRounds {
// 			// #TODO change delay logic
// 			go func() {
// 				time.Sleep(2 * time.Second)

// 				send_QwithCriteria(s, i.ChannelID, db, "session", cat, lvl)
// 			}()
// 		} else {
// 			finishSession(s, i.ChannelID, session)
// 		}
// 	} else {

// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{Content: "❌ خطأ! حاول مرة أخرى.", Flags: discordgo.MessageFlagsEphemeral},
// 		})
// 	}
// }

func finishSession(s *discordgo.Session, channelID string, session *QuizSession) {
	sessionsMu.Lock()
	delete(activeSessions, channelID)
	sessionsMu.Unlock()

	leaderboard := "🏁 **انتهى الكويز! النتائج النهائية:**\n"
	for id, score := range session.Scores {
		leaderboard += fmt.Sprintf("👤 %s: %d نقطة\n", session.Participants[id], score)
	}

	s.ChannelMessageSend(channelID, leaderboard)
}

func startSessionWithParams(s *discordgo.Session, channID, userID string, db *database.Queries, cat []int, diff int) {
	sessionsMu.Lock()
	if _, exists := activeSessions[channID]; exists {
		sessionsMu.Unlock()
		s.ChannelMessageSend(channID, "⚠️ هناك كويز جاري بالفعل في هذه القناة!")
		return
	}
	session := &QuizSession{
		ChannelID:    channID,
		StartedBy:    userID,
		MaxRounds:    5,
		CategoryID:   cat,
		Difficulty:   diff,
		Scores:       make(map[string]int),
		Participants: make(map[string]string),
		IsActive:     true,
	}
	activeSessions[channID] = session
	sessionsMu.Unlock()

	s.ChannelMessageSend(channID, "🎮 **بدأ الكويز!** استعدوا لـ 5 جولات...")
	send_QwithCriteria(s, channID, db, "session", cat, diff)
}

func Start_session(s *discordgo.Session, m *discordgo.MessageCreate, db *database.Queries) {
	sessionsMu.RLock()
	_, exists := activeSessions[m.ChannelID]
	sessionsMu.RUnlock()
	pendingMu.Lock()
	pendingSetups[m.ChannelID] = &pendingSetup{StartedBy: m.Author.ID}
	pendingMu.Unlock()
	if exists {
		s.ChannelMessageSend(m.ChannelID, "⚠️ هناك كويز جاري بالفعل في هذه القناة!")
		return
	}
	SendSetupMessage(s, m.ChannelID)
}

// StartSetupFromSlash is the slash-command entry point for /quiz.
func StartSetupFromSlash(s *discordgo.Session, channID, userID string, db *database.Queries) {
	sessionsMu.RLock()
	_, exists := activeSessions[channID]
	sessionsMu.RUnlock()
	if exists {
		s.ChannelMessageSend(channID, "⚠️ هناك كويز جاري بالفعل في هذه القناة!")
		return
	}
	pendingMu.Lock()
	pendingSetups[channID] = &pendingSetup{StartedBy: userID}
	pendingMu.Unlock()
	slog.Info("slash /quiz triggered", "channel", channID, "user", userID)
	SendSetupMessage(s, channID)
}

// GetQFromSlash is the slash-command entry point for /سؤال.
func GetQFromSlash(s *discordgo.Session, channID string, db *database.Queries, cat, lvl int) {
	if cat == 0 {
		cat = rand.IntN(6) + 1
	}
	if lvl == 0 {
		lvl = rand.IntN(3) + 1
	}
	slog.Info("slash /سؤال", "channel", channID, "cat", cat, "lvl", lvl)
	send_QwithCriteria(s, channID, db, "quiz", []int{cat}, lvl)
}
