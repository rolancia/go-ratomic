package ratomic

import (
	"fmt"
)

type LockKey string

type LockKeyPrefix string

func (kPre LockKeyPrefix) Merge(key LockKey) LockKey {
	return LockKey(fmt.Sprintf("%s.%s", kPre, key))
}

type Driver interface {
	KeyPrefix() LockKeyPrefix
	MSetNX(keys []LockKey) (int64, *DriverError)
	MDel(keys []LockKey) (int64, *DriverError)
}
