package timewatch

import (
	"time"
)

type Cache interface {
	HGetAll(k string) (result map[string]string, err error)
	HGet(k, field string) (string, error)
	HSet(k string, fields ...string) error
	HDel(k string, field string) error
	SetNX(k, v string, expiration time.Duration) (bool, error)
	Del(key string) error
}
