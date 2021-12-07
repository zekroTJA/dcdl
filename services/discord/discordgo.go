package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/zekrotja/dcdl/commands"
	"github.com/zekrotja/dcdl/services/config"
	"github.com/zekrotja/dcdl/services/storage"
	"github.com/zekrotja/dcdl/static"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/middlewares/ratelimit/v2"
	"github.com/zekrotja/ken/state"
	"github.com/zekrotja/ken/store"
)

const intents = discordgo.IntentsGuilds

type DiscordGo struct {
	*discordgo.Session

	cfg config.ConfigProvider
	ken *ken.Ken
}

var _ DiscordProvider = (*DiscordGo)(nil)

func NewDiscordGo(cfg config.ConfigProvider, st storage.StorageProvider) (dg *DiscordGo, err error) {
	dg = new(DiscordGo)
	dg.cfg = cfg

	dg.Session, err = discordgo.New("Bot " + cfg.Instance().Discord.Token)
	if err != nil {
		return
	}

	dg.StateEnabled = true
	dg.Identify.Intents = discordgo.MakeIntent(intents)

	dg.AddHandler(dg.readyHandler)

	dg.ken, err = ken.New(dg.Session, ken.Options{
		CommandStore:   store.NewDefault(),
		State:          state.NewInternal(),
		OnSystemError:  systemErrorHandler,
		OnCommandError: commandErrorHandler,
		EmbedColors: ken.EmbedColors{
			Default: static.ColorMain,
			Error:   static.ColorError,
		},
	})
	if err != nil {
		return
	}

	err = dg.ken.RegisterMiddlewares(
		ratelimit.New(),
	)
	if err != nil {
		return
	}

	err = dg.ken.RegisterCommands(
		&commands.Collect{st, cfg},
		&commands.Info{},
	)

	return
}

func (dg *DiscordGo) Run() (err error) {
	err = dg.Open()
	return
}

func (dg *DiscordGo) readyHandler(_ *discordgo.Session, e *discordgo.Ready) {
	logrus.WithField("uid", e.User.ID).WithField("uname", e.User.String()).Info("Discord session ready")
}

func systemErrorHandler(context string, err error, args ...interface{}) {
	logrus.WithField("ctx", context).WithError(err).Error("ken error")
}

func commandErrorHandler(err error, ctx *ken.Ctx) {
	// Is ignored if interaction has already been responded
	ctx.Defer()

	if err == ken.ErrNotDMCapable {
		ctx.FollowUpError("This command can not be used in DMs.", "")
		return
	}

	ctx.FollowUpError(
		fmt.Sprintf("The command execution failed unexpectedly:\n```\n%s\n```", err.Error()),
		"Command execution failed")
}
