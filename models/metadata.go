package models

import "github.com/bwmarrin/discordgo"

type AttMetadata struct {
	ArchiveFilename string `json:"archive_filename"`
	Hash            string `json:"hash"`
	IsDuplicate     bool   `json:"is_duplicate"`
	*discordgo.MessageAttachment
}

type MsgMetadata struct {
	GuildID     string        `json:"guild_id"`
	ChannelID   string        `json:"channel_id"`
	MessageID   string        `json:"message_id"`
	AuthorID    string        `json:"author_id"`
	Attachments []AttMetadata `json:"attachments"`
}
