package bot

import (
	"github.com/abdooman21/go-discord/quiz"
	"github.com/bwmarrin/discordgo"
)

func (api *Application) slashQuiz(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channID := i.ChannelID
	userID := i.Member.User.ID

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	quiz.StartSetupFromSlash(s, channID, userID, api.DB)

	s.InteractionResponseDelete(i.Interaction)
}

func (api *Application) slashSingleQuestion(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	cat := 0
	lvl := 0
	for _, opt := range options {
		switch opt.Name {
		case "category":
			cat = int(opt.IntValue())
		case "difficulty":
			lvl = int(opt.IntValue())
		}
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	quiz.GetQFromSlash(s, i.ChannelID, api.DB, cat, lvl)
	s.InteractionResponseDelete(i.Interaction)
}
