package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

type Info struct{}

var (
	_ ken.SlashCommand = (*Info)(nil)
)

func (c *Info) Name() string {
	return "info"
}

func (c *Info) Description() string {
	return "Display general help information about this bot."
}

func (c *Info) Version() string {
	return "1.0.0"
}

func (c *Info) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Info) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (c *Info) Run(ctx *ken.Ctx) (err error) {
	ctx.Session.State.User.AvatarURL("64x64")
	err = ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title: "Information",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    ctx.Session.State.User.AvatarURL("64x64"),
			Width:  64,
			Height: 64,
		},
		Description: "This bot can be used to download message attachments (Images, Videos, ...) " +
			"from sent messages in a Discord channel.\n\n" +
			"Simply use the `/collect` slash command in the channel you want to collect images from and " +
			"the bot will start collecting all messages, analyzing them for attachments and then downloading them " +
			"to the bot's storage. Afetr that, you will receive a link to download an archive with the attachments.\n\n" +
			"For more information, read the [**Readme**](https://github.com/zekroTJA/dcdl/blob/master/README.md) on GitHub.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Useful Links",
				Value: "- [GitHub Repository](https://github.com/zekroTJA/dcdl)\n" +
					"- [Issue Tracker](https://github.com/zekroTJA/dcdl/issues)\n" +
					"- [Creator (zekro)](https://www.zekro.de)",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "© 2021 Ringo Hoffmann (zekro Development)",
		},
	})
	return
}
