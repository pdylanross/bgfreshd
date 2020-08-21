package db

import (
	bolt "go.etcd.io/bbolt"
	"strconv"
	"time"
)

type sourceDb struct {
	sourceUniqueName string
	parent *backgroundDb
}

func (s *sourceDb) SetString(key string, val string) error {
	return s.parent.db.Update(func(tx *bolt.Tx) error {
		var bucket = tx.Bucket([]byte(s.sourceUniqueName))
		return s.parent.putString(bucket, key, val)
	})
}

func (s *sourceDb) GetString(key string) (string, error) {
	tx, err := s.parent.db.Begin(false)
	if err != nil {
		return "", err
	}

	defer tx.Rollback()
	bucket := tx.Bucket([]byte(s.sourceUniqueName))
	return s.parent.getString(bucket, key)
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
		bucket := tx.Bucket([]byte(s.sourceUniqueName))
		return s.parent.putBool(bucket, key, val)
	})
}

func (s *sourceDb) GetBool(key string) (bool, error) {
	tx, err := s.parent.db.Begin(false)
	if err != nil {
		return false, err
	}

	bucket := tx.Bucket([]byte(s.sourceUniqueName))
	return s.parent.getBool(bucket, key)
}

func (s *sourceDb) KeyExists(key string) (bool, error) {
	tx, err := s.parent.db.Begin(false)
	if err != nil {
		return false, err
	}

	bucket := tx.Bucket([]byte(s.sourceUniqueName))
	return bucket.Get([]byte(key)) != nil, nil
}

