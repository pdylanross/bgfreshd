package db

import (
	bolt "go.etcd.io/bbolt"
	"strconv"
	"time"
)

func (b *backgroundDb) getBucket(tx *bolt.Tx, key string) *bolt.Bucket {
	return tx.Bucket([]byte(key))
}

func (b *backgroundDb) putString(bucket *bolt.Bucket, key string, val string) error {
	return bucket.Put([]byte(key), []byte(val))
}

func (b *backgroundDb) getString(bucket *bolt.Bucket, key string) string {
	val := bucket.Get([]byte(key))
	return string(val)
}

func (b *backgroundDb) putTime(bucket *bolt.Bucket, key string, val time.Time) error {
	formatted := strconv.FormatInt(val.UTC().Unix(), 10)
	return b.putString(bucket, key, formatted)
}

func (b *backgroundDb) getTime(bucket *bolt.Bucket, key string) (time.Time, error) {
	val := b.getString(bucket, key)
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(parsed, 0), nil
}

func (b *backgroundDb) putBool(bucket *bolt.Bucket, key string, val bool) error {
	bytes := []byte{0}
	if val {
		bytes[0] = 1
	}

	return bucket.Put([]byte(key), bytes)
}

func (b *backgroundDb) getBool(bucket *bolt.Bucket, key string) bool {
	val := bucket.Get([]byte(key))
	return len(val) == 1 && val[0] == 1
}
