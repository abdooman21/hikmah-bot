package bot

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// allCommands defines every slash command the bot registers.
var allCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "quiz",
		Description: "ابدأ جلسة كويز تفاعلية",
	},
	{
		Name:        "سؤال",
		Description: "احصل على سؤال واحد عشوائي",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "category",
				Description: "التصنيف (1-6)",
				Required:    false,
				MinValue:    func() *float64 { v := 1.0; return &v }(),
				MaxValue:    6,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "difficulty",
				Description: "الصعوبة (1-3)",
				Required:    false,
				MinValue:    func() *float64 { v := 1.0; return &v }(),
				MaxValue:    3,
			},
		},
	},
	{
		Name:        "راديو",
		Description: "شغّل إذاعة إسلامية في القناة الصوتية",
	},
	{
		Name:        "وقف",
		Description: "أوقف البث وأخرج من القناة الصوتية",
	},
	{
		Name:        "مساعدة",
		Description: "اعرض قائمة الأوامر المتاحة",
	},
}

func (api *Application) RegisterCommands(guildID string) {
	for _, cmd := range allCommands {
		_, err := api.Bot.ApplicationCommandCreate(api.Bot.State.User.ID, guildID, cmd)
		if err != nil {
			slog.Error("failed to register command", "cmd", cmd.Name, "err", err)
		} else {
			slog.Info("registered command", "cmd", cmd.Name)
		}
	}
}

func (api *Application) UnregisterCommands(guildID string) {
	cmds, err := api.Bot.ApplicationCommands(api.Bot.State.User.ID, guildID)
	if err != nil {
		slog.Error("failed to fetch commands", "err", err)
		return
	}
	for _, cmd := range cmds {
		api.Bot.ApplicationCommandDelete(api.Bot.State.User.ID, guildID, cmd.ID)
		slog.Info("deleted command", "cmd", cmd.Name)
	}
}
