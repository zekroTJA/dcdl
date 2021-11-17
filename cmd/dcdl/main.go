package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/zekroTJA/shinpuru/pkg/argp"
	"github.com/zekrotja/dcdl/services/config"
	"github.com/zekrotja/dcdl/services/discord"
	"github.com/zekrotja/dcdl/services/storage"
	"github.com/zekrotja/dcdl/services/webserver"
)

func main() {
	godotenv.Load()

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05 MST",
	})

	flagConfig, _ := argp.String("-c", "config.yml", "Optional config file location.")

	if flagHelp, _ := argp.Bool("-h", false, "Display help."); flagHelp {
		fmt.Println("Usage:\n" + argp.Help())
		return
	}

	var err error

	cfg := config.NewPaerser(argp.Args(), flagConfig)

	logrus.Info("Loading config ...")
	if err = cfg.Load(); err != nil {
		logrus.WithError(err).Fatal("Failed loading config")
	}

	st := storage.NewLocal(cfg)

	logrus.Info("Staring Discord bot session ...")
	bot, err := discord.NewDiscordGo(cfg, st)
	if err = bot.Open(); err != nil {
		logrus.WithError(err).Fatal("Failed opening Discord session")
	}
	defer bot.Close()

	logrus.Info("Starting web server ...")
	ws := webserver.NewFiberServer(cfg, st)
	go func() {
		if err = ws.Run(); err != nil {
			logrus.WithError(err).Fatal("Failed setting up web server")
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	logrus.Info("Shutting down ...")
}
