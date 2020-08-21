package main

import (
	"bgfreshd/internal"
	"bgfreshd/internal/config"
	"bgfreshd/internal/db"
	_ "bgfreshd/internal/filters"
	"bgfreshd/internal/pipeline"
	_ "bgfreshd/internal/sources"
	"bgfreshd/pkg/background"
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

type Service interface {
	init() error
	Run() error
	Stop()
}

type bgFreshService struct {
	config            *config.Config
	pipeline          pipeline.Pipeline
	db                db.BackgroundDb
	logger            *logrus.Entry
	ctx               context.Context
	cancel            context.CancelFunc
	triggerGeneration chan int
}

func (b *bgFreshService) Stop() {
	b.db.Stop()
}

type BgFreshConfig struct {
	// Config is the file path to the configuration file
	Config string

	Log *logrus.Logger
}

func NewService(c *BgFreshConfig) (Service, error) {
	cfg, err := config.Load(c.Config)
	if err != nil {
		return nil, err
	}

	pl, err := pipeline.NewPipeline(cfg, c.Log)
	if err != nil {
		return nil, err
	}

	d, err := db.NewDb(cfg, c.Log.WithFields(logrus.Fields{
		"section": "db",
	}))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	service := &bgFreshService{
		config:   cfg,
		pipeline: pl,
		db:       d,
		logger: c.Log.WithFields(logrus.Fields{
			"section": "service",
		}),
		ctx:               ctx,
		cancel:            cancel,
		triggerGeneration: make(chan int),
	}

	if err := service.init(); err != nil {
		return nil, err
	}

	return service, nil
}

func (b *bgFreshService) init() error {
	return nil
}

func (b *bgFreshService) Run() error {
	go backgroundJob(b, time.Second*30, watchOutput)
	for {
		stale, err := b.db.GetStaleBackgrounds()
		if err != nil {
			return err
		}

		if len(stale) != 0 {
			if err := b.refreshStale(stale); err != nil {
				return err
			}
		}

		active, err := b.db.GetActiveBackgrounds()
		if err != nil {
			return err
		}

		for len(active) < b.config.MaxBackgrounds {
			if err := b.loadOne(); err != nil {
				return err
			}

			active, err = b.db.GetActiveBackgrounds()
			if err != nil {
				return err
			}
		}
		runtime.GC()
		debug.FreeOSMemory()

		select {
		case <-time.After(time.Minute * 5):
		case <-b.triggerGeneration:
		case <-b.ctx.Done():
			return nil
		}
	}
}

func (b *bgFreshService) loadOne() error {
	bg, err := b.pipeline.LoadOne()
	if err != nil {
		return err
	}

	name := bg.GetName()
	b.logger.Infof("Background %s found", name)

	if !b.db.Exists(name) {
		if err := b.createImage(bg); err != nil {
			return err
		}
	} else {
		b.logger.Debugf("load attempt rejected for %s, already exists", name)
	}

	return nil
}

func (b *bgFreshService) createImage(bg background.Background) error {
	bg.SetActive()
	bg.GenerateExpiry(b.config.MinRotationAge, b.config.MaxRotationAge)
	if err := b.db.SaveMetadata(bg); err != nil {
		return err
	}

	filename := b.getFilename(bg.GetName())
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return os.ErrExist
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer internal.Deferrer(b.logger, file.Close)

	writer := bufio.NewWriter(file)
	if err := jpeg.Encode(writer, bg.GetImage(), &jpeg.Options{Quality: 100}); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (b *bgFreshService) refreshStale(stale []string) error {
	for _, val := range stale {
		if err := b.loadOne(); err != nil {
			return err
		}

		b.logger.Infof("marking %s inactive", val)
		if err := b.db.MarkInactive(val); err != nil {
			return err
		}

		b.logger.Infof("removing %s", val)
		filename := b.getFilename(val)
		if err := os.Remove(filename); err != nil {
			b.logger.Warnf("error removing %s: %s", val, err.Error())
		}
	}

	return nil
}

func (b *bgFreshService) getFilename(bgName string) string {
	return filepath.Join(b.config.OutputPath, fmt.Sprintf("%s.jpeg", bgName))
}

func backgroundJob(b *bgFreshService, occursEvery time.Duration, job func(b *bgFreshService)) {
	for {
		job(b)

		select {
		case <-time.After(occursEvery):
		case <-b.ctx.Done():
			return
		}
	}
}

func watchOutput(b *bgFreshService) {
	active, err := b.db.GetActiveBackgrounds()
	if err != nil {
		b.logger.Warnf("watch output get active error: %s", err.Error())
	}

	anyMarked := false
	for _, name := range active {
		filename := b.getFilename(name)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			b.logger.Infof("background %s file missing - marking stale", name)
			err := b.db.MarkStale(name)
			anyMarked = true
			if err != nil {
				b.logger.Warnf("watch output mark stale error: %s", err.Error())
				continue
			}
		}
	}

	if anyMarked {
		b.triggerGeneration <- 1
	}
}
