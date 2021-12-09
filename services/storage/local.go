package storage

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/valyala/fasthttp"
	"github.com/zekroTJA/timedmap"
	"github.com/zekrotja/dcdl/models"
	"github.com/zekrotja/dcdl/services/config"
)

type Local struct {
	loc      string
	lifeTime time.Duration

	storeList *timedmap.TimedMap
}

var _ StorageProvider = (*Local)(nil)

func NewLocal(cfg config.ConfigProvider) (s *Local) {
	s = new(Local)
	s.loc = cfg.Instance().Storage.Location
	s.lifeTime = time.Duration(cfg.Instance().Storage.LifetimeSeconds) * time.Second
	s.storeList = timedmap.New(5 * time.Minute)

	return
}

func (s *Local) Store(
	id string,
	msgs []*discordgo.Message,
	includeMetadata,
	includeFiles bool,
	cStatus chan<- *discordgo.MessageAttachment,
) (err error) {
	file, err := os.Create(s.floc(id))
	if err != nil {
		return
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	if includeFiles {
		for _, msg := range msgs {
			for _, att := range msg.Attachments {
				if cStatus != nil {
					cStatus <- att
				}
				fName := fmt.Sprintf("%s-%s", msg.ID, att.Filename)
				var w io.Writer
				w, err = zw.Create(path.Join("files", fName))
				if err != nil {
					return
				}
				err = getFileFromUrl(att.URL, w)
				if err != nil {
					return
				}
			}
		}
	}

	if includeMetadata {
		metas := make([]models.MsgMetadata, len(msgs))
		for i, msg := range msgs {
			m := models.MsgMetadata{
				GuildID:     msg.GuildID,
				ChannelID:   msg.ChannelID,
				MessageID:   msg.ID,
				AuthorID:    msg.Author.ID,
				Attachments: make([]models.AttMetadata, len(msg.Attachments)),
			}
			for j, att := range msg.Attachments {
				m.Attachments[j] = models.AttMetadata{
					ArchiveFilename:   fmt.Sprintf("%s-%s", msg.ID, att.Filename),
					MessageAttachment: att,
				}
			}
			metas[i] = m
		}

		var w io.Writer
		w, err = zw.Create("metadata.json")
		if err != nil {
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err = enc.Encode(metas); err != nil {
			return
		}
	}

	s.storeList.Set(id, false, s.lifeTime, func(_ interface{}) {
		s.Delete(id)
	})

	return
}

func (s *Local) Get(id string) (r io.ReadCloser, size int64, err error) {
	_, ok := s.storeList.GetValue(id).(bool)
	if !ok {
		err = ErrNotFound
		return
	}

	fName := s.floc(id)

	i, err := os.Stat(fName)
	if err != nil {
		err = wrapNotFound(err)
		return
	}

	size = i.Size()

	r, err = os.Open(fName)
	if err != nil {
		return
	}

	return
}

func (s *Local) Delete(id string) (err error) {
	err = os.Remove(s.floc(id))
	err = wrapNotFound(err)
	return
}

func (s *Local) floc(id string, att ...string) string {
	p := []string{s.loc, id}
	return path.Join(append(p, att...)...)
}

func getFileFromUrl(url string, target io.Writer) (err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetMethod("GET")
	req.SetRequestURI(url)
	if err = fasthttp.Do(req, res); err != nil {
		return
	}

	if code := res.StatusCode(); code != 200 {
		err = fmt.Errorf("request response failed (%d)", code)
		return
	}

	err = res.BodyWriteTo(target)
	return
}

func wrapNotFound(err error) error {
	if err == os.ErrNotExist {
		return ErrNotFound
	}
	return err
}
