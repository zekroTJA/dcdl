package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/xid"
	"github.com/zekroTJA/shinpuru/pkg/bytecount"
	"github.com/zekrotja/dcdl/pkg/accmsgutil"
	"github.com/zekrotja/dcdl/services/config"
	"github.com/zekrotja/dcdl/services/storage"
	"github.com/zekrotja/dcdl/static"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/middlewares/ratelimit/v2"
)

const askForDownloadSize = 1 * 1024 * 1024 * 1024 // 1GiB

type Collect struct {
	St  storage.StorageProvider
	Cfg config.ConfigProvider
}

var (
	_ ken.SlashCommand         = (*Collect)(nil)
	_ ratelimit.LimitedCommand = (*Collect)(nil)
)

func (c *Collect) Name() string {
	return "collect"
}

func (c *Collect) Description() string {
	return "Collect attachments from a channel."
}

func (c *Collect) Version() string {
	return "1.1.0"
}

func (c *Collect) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (c *Collect) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type: discordgo.ApplicationCommandOptionChannel,
			Name: "channel",
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Description: "The channel to be analyzed (current if unset).",
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "limit",
			Description: "Limit the amount messages to be analyzed for collecting attachments (all if unset).",
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "include-metadata",
			Description: "Include message and attachment metadata in extra file (true if unset).",
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "include-files",
			Description: "Include attachment files (true in unset).",
		},
	}
}

func (c *Collect) LimiterBurst() int {
	return 1
}

func (c *Collect) LimiterRestoration() time.Duration {
	return time.Duration(c.Cfg.Instance().Discord.CooldownSeconds) * time.Second
}

func (c *Collect) IsLimiterGlobal() bool {
	return true
}

func (c *Collect) Run(ctx *ken.Ctx) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	ch, err := ctx.Channel()
	if err != nil {
		return
	}
	if v, ok := ctx.Options().GetByNameOptional("channel"); ok {
		ch = v.ChannelValue(ctx)
	}

	limit := 0
	if v, ok := ctx.Options().GetByNameOptional("limit"); ok {
		limit = int(v.IntValue())
	}

	totalLimit := int(c.Cfg.Instance().Discord.MessageLimit)

	if limit < 0 || (limit > totalLimit && totalLimit != 0) {
		add := ""
		if totalLimit != 0 {
			add = fmt.Sprintf(" (but not larger than `%d`)", totalLimit)
		}
		err = ctx.FollowUpError(
			"Limit must be either `0` (equals unlimited) or a value larger than `0`"+add+".",
			"Argument Error").Error
		ratelimit.Skip(ctx)
		return
	}

	includeMetadata := true
	if v, ok := ctx.Options().GetByNameOptional("include-metadata"); ok {
		includeMetadata = v.BoolValue()
	}

	includeFiles := true
	if v, ok := ctx.Options().GetByNameOptional("include-files"); ok {
		includeFiles = v.BoolValue()
	}

	if !(includeMetadata || includeFiles) {
		err = ctx.FollowUpError(
			"Either `include-metadata` or `include-files` must be set to `true`.",
			"Argument Error").Error
		ratelimit.Skip(ctx)
		return
	}

	curr := 0
	fum := ctx.FollowUpEmbed(collectMessagesEmbed(curr, limit))
	if err = fum.Error; err != nil {
		return
	}

	allMsgs := make([]*discordgo.Message, 0)
	var msgs []*discordgo.Message
	var lastMsgID string

	for {
		msgs, err = ctx.Session.ChannelMessages(ch.ID, 100, lastMsgID, "", "")
		if err != nil {
			return
		}
		if len(msgs) <= 0 {
			break
		}

		curr += len(msgs)
		lastMsgID = msgs[len(msgs)-1].ID

		for _, msg := range msgs {
			if len(msg.Attachments) != 0 {
				allMsgs = append(allMsgs, msg)
			}
		}

		if err = fum.EditEmbed(collectMessagesEmbed(curr, limit)); err != nil {
			return
		}

		if curr >= limit && limit != 0 || curr >= totalLimit && totalLimit != 0 {
			break
		}
	}

	attCount := 0
	attTotalSize := 0
	for _, msg := range allMsgs {
		attCount += len(msg.Attachments)
		for _, att := range msg.Attachments {
			attTotalSize += att.Size
		}
	}

	if attCount == 0 {
		err = fum.EditEmbed(&discordgo.MessageEmbed{
			Color:       static.ColorError,
			Description: "Collected messages do not contain any attachments.",
		})
		return
	}

	maxTotalSize := c.Cfg.Instance().Discord.SizeLimitU
	if includeFiles {
		var ok bool
		if uint64(attTotalSize) > maxTotalSize {
			ok, err = accmsgutil.Wrap(ctx, fmt.Sprintf(
				"Total attachment size (`%s`) exceed configured maximum attachment size (`%s`).\n\n"+
					"Do you want to continue with only the metadata file?",
				bytecount.Format(uint64(attTotalSize)), bytecount.Format(maxTotalSize)))
			if err != nil {
				return
			}
			if !ok {
				err = fum.EditEmbed(&discordgo.MessageEmbed{
					Color:       static.ColorError,
					Description: "Canceled.",
				})
				return
			}
			includeFiles = false
		} else if uint64(attTotalSize) >= askForDownloadSize {
			ok, err = accmsgutil.Wrap(ctx, fmt.Sprintf(
				"Total attachment size (`%s`) is fairly large and might take some while to download.\n\n"+
					"Do you rather just want to download the metadata file and download the atatchments yourself?\n"+
					"*(Cancel to proceed with attachment download)*",
				bytecount.Format(uint64(attTotalSize))))
			if err != nil {
				return
			}
			includeFiles = !ok
		}
	}

	var emb *discordgo.MessageEmbed
	if includeFiles {
		emb = processEmbed(fmt.Sprintf(
			"Downloading attachments of collected messages (`%d` attachments) ...\n"+
				"*This can take some time depending on the size of the attachments.*",
			attCount))
	} else {
		emb = processEmbed(fmt.Sprintf(
			"Assembling attachment metadata (`%d` attachments) ...\n",
			attCount))
	}
	if err = fum.EditEmbed(emb); err != nil {
		return
	}

	id := fmt.Sprintf("%s-%s-%s", ctx.Event.GuildID, ch.ID, xid.New().String())

	var cStatus chan *discordgo.MessageAttachment
	if includeFiles {
		cStatus = make(chan *discordgo.MessageAttachment)
		go func() {
			i := 0
			for att := range cStatus {
				i++
				emb := processEmbed(fmt.Sprintf(
					"Downloading attachments of collected messages (`%d`/`%d` attachments) ...\n"+
						"*This can take some time depending on the size of the attachments.*\n\n"+
						"Downloading `%s` (%s)...",
					i, attCount, att.Filename, bytecount.Format(uint64(att.Size))))
				fum.EditEmbed(emb)
			}
		}()
	}

	err = c.St.Store(id, allMsgs, includeMetadata, includeFiles, cStatus)
	if err != nil {
		return
	}

	dlUrl := fmt.Sprintf("%s/collections/%s.zip",
		c.Cfg.Instance().Webserver.PublicAddress, id)
	timeout := (time.Duration(c.Cfg.Instance().Storage.LifetimeSeconds) * time.Second).Round(time.Second).String()
	emb = &discordgo.MessageEmbed{
		Color: static.ColorMain,
		Description: fmt.Sprintf(
			"You can now download the collection archive using the following link.\n%s"+
				"\n\n*This link will expire in %s.*", dlUrl, timeout),
	}

	fallback := func() error {
		return fum.EditEmbed(emb)
	}

	dmCh, err := ctx.Session.UserChannelCreate(ctx.User().ID)
	if err != nil {
		return fallback()
	}
	if _, err = ctx.Session.ChannelMessageSendEmbed(dmCh.ID, emb); err != nil {
		return fallback()
	}

	err = fum.EditEmbed(&discordgo.MessageEmbed{
		Color:       static.ColorMain,
		Description: "I've sent you a DM with the download link to the collection archive. 😉",
	})

	return
}

func collectMessagesEmbed(curr, limit int) *discordgo.MessageEmbed {
	if limit != 0 {
		return processEmbed(fmt.Sprintf(
			"Collecting messages (`%d` of `%d`) ...",
			curr, limit))
	}
	return processEmbed(fmt.Sprintf(
		"Collecting messages (`%d`) ...",
		curr))
}

func processEmbed(content string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:       static.ColorPending,
		Description: ":clock10: " + content,
	}
}
