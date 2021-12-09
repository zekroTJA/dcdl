package storage

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/dcdl/models"
)

type StorageProvider interface {
	Store(
		id string,
		msgs []*discordgo.Message,
		includeMetadata,
		includeFiles bool,
		excludeDuplicates bool,
		cStatus chan<- *models.AttMetadata,
	) (err error)
	Get(id string) (r io.ReadCloser, size int64, err error)
	Delete(id string) (err error)
}
