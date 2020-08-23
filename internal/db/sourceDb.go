package db

import (
	"bgfreshd/pkg"
	bolt "go.etcd.io/bbolt"
	"strconv"
	"time"
)

type sourceDb struct {
	sourceUniqueName string
	parent           *backgroundDb
}

func (s *sourceDb) SetString(key string, val string) error {
	return s.parent.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(s.sourceUniqueName))
		if err != nil {
			return err
		}

		return s.parent.putString(bucket, key, val)
	})
}

func (s *sourceDb) GetString(key string) (string, error) {
	ret := ""
	err := s.parent.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.sourceUniqueName))
		if bucket == nil {
			return pkg.NotFoundError{Key: key}
		}

		var err error
		ret, err = s.parent.getString(bucket, key)
		return err
	})
	return ret, err
}

func (s *sourceDb) SetTime(key string, time time.Time) error {
	formatted := strconv.FormatInt(time.UTC().Unix(), 10)
	return s.SetString(key, formatted)
}

func (s *sourceDb) GetTime(key string) (time.Time, error) {
	str, err := s.GetString(key)
	if err != nil {
		return time.Time{}, err
	}

	parsed, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(parsed, 0), nil
}

func (s *sourceDb) SetBool(key string, val bool) error {
	return s.parent.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(s.sourceUniqueName))
		if err != nil {
			return err
		}

		return s.parent.putBool(bucket, key, val)
	})
}

func (s *sourceDb) GetBool(key string) (bool, error) {
	ret := false
	err := s.parent.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.sourceUniqueName))
		if bucket == nil {
			return &pkg.NotFoundError{Key: key}
		}

		var err error
		ret, err = s.parent.getBool(bucket, key)
		return err
	})
	return ret, err
}

func (s *sourceDb) KeyExists(key string) (bool, error) {
	ret := false
	err := s.parent.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.sourceUniqueName))
		if bucket == nil {
			return nil
		}

		ret = bucket.Get([]byte(key)) != nil
		return nil
	})
	return ret, err
}
