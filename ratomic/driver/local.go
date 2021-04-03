package driver

import (
	"github.com/rolancia/go-ratomic/ratomic"
	"sync"
	"time"
)

func NewLocalDriver(keyPrefix ratomic.LockKeyPrefix, networkLatency time.Duration) *LocalDriver {
	return &LocalDriver{
		NetworkLatency: networkLatency,

		keyPrefix: keyPrefix,
		locks:     map[ratomic.LockKey]bool{},
	}
}

type LocalDriver struct {
	NetworkLatency time.Duration

	keyPrefix ratomic.LockKeyPrefix

	locks   map[ratomic.LockKey]bool
	muLocks sync.Mutex
}

func (dri *LocalDriver) KeyPrefix() ratomic.LockKeyPrefix {
	return dri.keyPrefix
}

// to get lock. dummy of Redis SetNX
func (dri *LocalDriver) MSetNX(keys []ratomic.LockKey) (int64, *ratomic.DriverError) {
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
func (dri *LocalDriver) MDel(keys []ratomic.LockKey) (int64, *ratomic.DriverError) {
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
