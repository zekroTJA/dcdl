package accmsgutil

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/pkg/acceptmsg"
	"github.com/zekrotja/ken"
)

func Wrap(ctx *ken.Ctx, msg string) (accepted bool, err error) {
	var am *acceptmsg.AcceptMessage
	am, err = acceptmsg.New().
		WithSession(ctx.Session).
		DeleteAfterAnswer().
		LockOnUser(ctx.User().ID).
		WithContent(msg).
		DoOnAccept(func(m *discordgo.Message) error {
			accepted = true
			return nil
		}).
		DoOnDecline(func(m *discordgo.Message) error {
			return nil
		}).
		AsFollowUp(ctx)
	if err != nil {
		return
	}
	err = am.Error()
	return
}
