package storage

import (
	"io"

	"github.com/bwmarrin/discordgo"
)

type StorageProvider interface {
	Store(
		id string,
		msgs []*discordgo.Message,
		includeMetadata,
		includeFiles bool,
		cStatus chan<- *discordgo.MessageAttachment,
	) (err error)
	Get(id string) (r io.ReadCloser, size int64, err error)
	Delete(id string) (err error)
}
