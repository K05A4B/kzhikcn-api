package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
)

type BadgerCache struct {
	db *badger.DB
}

func (c *BadgerCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), val)
		if ttl > 0 {
			e = e.WithTTL(ttl)
		}
		return txn.SetEntry(e)
	})
}

func (c *BadgerCache) Get(ctx context.Context, key string) ([]byte, error) {
	var result []byte

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			result = append([]byte{}, val...)
			return nil
		})
	})

	return result, err
}

func (c *BadgerCache) Delete(ctx context.Context, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (c *BadgerCache) Exists(ctx context.Context, key string) (bool, error) {
	val, err := c.Get(ctx, key)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return val != nil, nil
}
