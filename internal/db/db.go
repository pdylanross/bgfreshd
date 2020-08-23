package db

import (
	"bgfreshd/internal/config"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/source"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"time"
)

type BackgroundDb interface {
	Stop()
	SaveMetadata(bg background.Background) error
	GetBackgroundList() ([]string, error)
	GetActiveBackgrounds() ([]string, error)
	GetStaleBackgrounds() ([]string, error)
	MarkInactive(name string) error
	MarkStale(name string) error
	Exists(key string) bool
	NewSourceDb(sourceName string, sourceMeta string) (source.Db, error)
}

type backgroundDb struct {
	config *config.Config
	logger *logrus.Entry
	db     *bolt.DB
	ctx    context.Context
	cancel context.CancelFunc
}

func (b *backgroundDb) NewSourceDb(sourceName string, sourceMeta string) (source.Db, error) {
	uniq := fmt.Sprintf("%s-%s", sourceName, sourceMeta)
	return &sourceDb{
		sourceUniqueName: uniq,
		parent:           b,
	}, nil
}

func (b *backgroundDb) Stop() {
	b.logger.Info("Db Shutdown")
	b.cancel()
}

func (b *backgroundDb) SaveMetadata(bg background.Background) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket, err := tx.CreateBucketIfNotExists([]byte(bg.GetName()))
	if err != nil {
		return err
	}

	for _, metaKey := range bg.GetMetadataKeys() {
		if err := b.putString(bucket, metaKey, bg.GetMetadata(metaKey)); err != nil {
			return err
		}
	}

	if err := b.putString(bucket, "name", bg.GetName()); err != nil {
		return err
	}
	if err := b.putTime(bucket, "created_date", bg.GetCreatedDate()); err != nil {
		return err
	}
	if err := b.putTime(bucket, "expires_on", bg.GetExpiry()); err != nil {
		return err
	}
	if err := b.putBool(bucket, "active", bg.IsActive()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (b *backgroundDb) GetBackgroundList() ([]string, error) {
	var names []string
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			names = append(names, string(name))
			return nil
		})
	})
	return names, err
}

func (b *backgroundDb) Exists(key string) bool {
	exists := false
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(key))
		exists = bucket != nil
		return nil
	})
	if err != nil {
		b.logger.Warnf("Error checking bucket existence: %s", err)
	}

	return err == nil && exists
}

func (b *backgroundDb) GetActiveBackgrounds() ([]string, error) {
	var active []string
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			isActive, _ := b.getBool(bucket, "active")
			if isActive {
				active = append(active, string(name))
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return active, nil
}

func (b *backgroundDb) GetStaleBackgrounds() ([]string, error) {
	var stale []string
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			isStale, _ := b.getBool(bucket, "stale")
			isActive, _ := b.getBool(bucket, "active")
			if isActive && isStale {
				stale = append(stale, string(name))
			}
			return nil
		})
	})
	return stale, err
}

func (b *backgroundDb) MarkInactive(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		return b.putBool(bucket, "active", false)
	})
}

func (b *backgroundDb) MarkStale(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		return b.putBool(bucket, "stale", true)
	})
}

func NewDb(config *config.Config, logger *logrus.Entry) (BackgroundDb, error) {
	outputDir := config.OutputPath
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0777); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	dbFile := filepath.Join(config.OutputPath, ".bgfreshd.dat")
	dbOpts := &bolt.Options{
		Timeout: 5 * time.Second,
	}
	db, err := bolt.Open(dbFile, 0664, dbOpts)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	logger.Infof("Database loaded at %s", dbFile)
	ret := &backgroundDb{
		config: config,
		logger: logger,
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}

	go dbBackgroundJob(ret, 30*time.Minute, dbInfoPoll)
	go dbBackgroundJob(ret, 30*time.Minute, dbExpiryPoll)

	return ret, nil
}

func dbBackgroundJob(bgDb *backgroundDb, occursEvery time.Duration, job func(db *backgroundDb)) {
	for {
		job(bgDb)

		select {
		case <-time.After(occursEvery):
		case <-bgDb.ctx.Done():
			return
		}
	}
}

func dbInfoPoll(b *backgroundDb) {
	stats := b.db.Stats()
	b.logger.Debugf("FreePageN: %d PendingPageN: %d FreeAlloc: %d FreelistInuse: %d TxN: %d OpenTxN: %d", stats.FreePageN, stats.PendingPageN, stats.FreeAlloc, stats.FreelistInuse, stats.TxN, stats.OpenTxN)
}

func dbExpiryPoll(b *backgroundDb) {
	b.logger.Debug("expiry poll start")
	active, err := b.GetActiveBackgrounds()
	if err != nil {
		b.logger.Warnf("expiry poll error: %s", err.Error())
	}
	now := time.Now()
	for _, val := range active {
		err := b.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(val))
			expiry, err := b.getTime(bucket, "expires_on")
			if err != nil {
				return err
			}

			if now.After(expiry) {
				b.logger.Infof("marking %s stale", val)
				return b.putBool(bucket, "stale", true)
			}

			return nil
		})

		if err != nil {
			b.logger.Warnf("error marking expiry: %s", err.Error())
		}
	}
}
