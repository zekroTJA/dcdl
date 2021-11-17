package discord

type DiscordProvider interface {
	Run() (err error)
	Close() (err error)
}
