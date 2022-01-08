package webserver

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zekrotja/dcdl/services/config"
	"github.com/zekrotja/dcdl/services/storage"
)

type FiberServer struct {
	*fiber.App

	cfg config.ConfigProvider
	st  storage.StorageProvider
}

var _ WebserverProvider = (*FiberServer)(nil)

func NewFiberServer(cfg config.ConfigProvider, st storage.StorageProvider) (s *FiberServer) {
	s = new(FiberServer)
	s.cfg = cfg
	s.st = st

	s.App = fiber.New(fiber.Config{
		ServerHeader:          "dcdl",
		GETOnly:               true,
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-For",
	})

	s.Use(Logger())

	s.Get("/collections/:id", s.getFileHandler)

	return
}

func (s *FiberServer) Run() (err error) {
	err = s.Listen(s.cfg.Instance().Webserver.BindAddress)
	return
}

func (s *FiberServer) getFileHandler(ctx *fiber.Ctx) (err error) {
	id := stripId(ctx.Params("id"))

	rc, size, err := s.st.Get(id)
	if err == storage.ErrNotFound {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	ctx.Response().Header.SetContentType("application/zip")
	return ctx.SendStream(rc, int(size))
}

func stripId(id string) string {
	split := strings.SplitN(id, ".", 2)
	return split[0]
}
