package ratomic

import (
	"fmt"
	"sync"
	"time"
)

type LockKeyPrefix string

func (kPre LockKeyPrefix) Merge(key string) string {
	return fmt.Sprintf("%s.%s", kPre, key)
}

type Driver interface {
	KeyPrefix() LockKeyPrefix
	MSetNX(keys ...string) (int64, *DriverError)
	Del(keys ...string) (int64, *DriverError)
}

//
func NewLocalDriver(keyPrefix LockKeyPrefix, networkLatency time.Duration) *LocalDriver {
	return &LocalDriver{
		NetworkLatency: networkLatency,

		keyPrefix: keyPrefix,
		locks:     map[string]bool{},
	}
}

type LocalDriver struct {
	NetworkLatency time.Duration

	keyPrefix LockKeyPrefix

	locks   map[string]bool
	muLocks sync.Mutex
}

func (dri *LocalDriver) KeyPrefix() LockKeyPrefix {
	return dri.keyPrefix
}

// to get lock. dummy of Redis MSetNX
func (dri *LocalDriver) MSetNX(keys ...string) (int64, *DriverError) {
	dri.waitForLatency()
	ok := true

	dri.muLocks.Lock()
	defer dri.muLocks.Unlock()

	for i := range keys {
		if _, exist := dri.locks[keys[i]]; exist {
			ok = false
			break
		}
	}

	if ok == false {
		return 0, nil
	}

	for i := range keys {
		dri.locks[keys[i]] = true
	}

	dri.waitForLatency()
	return 1, nil
}

// to release lock. dummy of Redis Del
func (dri *LocalDriver) Del(keys ...string) (int64, *DriverError) {
	dri.waitForLatency()
	var numDel int64 = 0

	dri.muLocks.Lock()
	defer dri.muLocks.Unlock()

	for i := range keys {
		if _, exist := dri.locks[keys[i]]; exist {
			delete(dri.locks, keys[i])
			numDel++
		}
	}
	dri.waitForLatency()

	return numDel, nil
}

func (dri *LocalDriver) waitForLatency() {
	if dri.NetworkLatency == 0 {
		return
	}
	time.Sleep(dri.NetworkLatency / 2)
}
