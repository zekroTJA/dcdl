package main

import (
	"encoding/json"
	"flag"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/zekroTJA/r34-crawler/pkg/workerpool"
	"github.com/zekroTJA/shinpuru/pkg/stringutil"
	"github.com/zekrotja/dcdl/models"
)

var (
	metaFile          string
	outputDir         string
	splitByUID        bool
	uidFiler          string
	parallelDownloads uint

	subPathCache []string
)

func parseFlags() {
	flag.StringVar(&metaFile, "i", "metadata.json", "The location of the metadata file.")
	flag.StringVar(&outputDir, "o", "files", "The output directory.")
	flag.BoolVar(&splitByUID, "split", false, "Split the downloaded files into seperate folders by author user ID.")
	flag.StringVar(&uidFiler, "filter", "", "Filter attachments to download by UIDs (comma separated list).")
	flag.UintVar(&parallelDownloads, "parallel", 2, "Parralell download count.")
	flag.Parse()
}

func must(err error, errMsg string) {
	if err != nil {
		logrus.WithError(err).Fatal(errMsg)
	}
}

func download(url, dest string) (err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetMethod("GET")
	req.SetRequestURI(url)

	if err = fasthttp.Do(req, res); err != nil {
		return
	}

	f, err := os.Create(dest)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(res.Body())

	return
}

func processAttachment(msg *models.MsgMetadata, att *models.AttMetadata) (err error) {
	var dest string
	if splitByUID {
		p := path.Join(outputDir, msg.AuthorID)
		if !stringutil.ContainsAny(msg.AuthorID, subPathCache) {
			if err = os.MkdirAll(p, os.ModeDir); err != nil {
				return
			}
			subPathCache = append(subPathCache, msg.AuthorID)
		}
		dest = path.Join(p, att.Filename)
	} else {
		dest = path.Join(outputDir, att.Filename)
	}

	// Skip if file already exists.
	_, err = os.Stat(dest)
	if !os.IsNotExist(err) {
		return
	}

	err = download(att.URL, dest)
	if err != nil {
		return
	}

	return
}

func main() {
	parseFlags()

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	err := os.MkdirAll(outputDir, os.ModeDir)
	must(err, "Failed to create output directory.")

	mf, err := os.Open(metaFile)
	must(err, "Failed to open meta file")
	defer mf.Close()

	var meta []*models.MsgMetadata
	err = json.NewDecoder(mf).Decode(&meta)
	must(err, "Failed to parse metadata file.")

	pool := workerpool.New(int(parallelDownloads))

	var successful, failed uint32
	go func() {
		for success := range pool.Results() {
			if success.(bool) {
				atomic.AddUint32(&successful, 1)
			} else {
				atomic.AddUint32(&failed, 1)
			}
		}
	}()

	uidFilterEnabled := len(uidFiler) != 0

	for _, msg := range meta {
		if uidFilterEnabled && !strings.Contains(uidFiler, msg.AuthorID) {
			continue
		}
		for _, att := range msg.Attachments {
			pool.Push(func(workerId int, params ...interface{}) interface{} {
				logrus.WithField("worker", workerId).WithField("name", att.Filename).Info("Download attachment")
				err := processAttachment(msg, &att)
				if err != nil {
					logrus.WithError(err).WithField("name", att.Filename).Error("Failed downloading file")
				}
				return err == nil
			})
		}
	}
	pool.Close()
	pool.WaitBlocking()

	// Wait until stats go routine has finished counting.
	time.Sleep(50 * time.Millisecond)

	logrus.WithFields(logrus.Fields{
		"successful": successful,
		"failed":     failed,
	}).Info("Download finished")
}
