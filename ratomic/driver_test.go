package ratomic

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewLocalDriver(t *testing.T) {
	ld := NewLocalDriver("lock", 3*time.Millisecond)

	key := LockKey("tae")

	for i := 0; i < 10; i++ {
		assert.False(t, isAlreadyLocked(ld.MSetNX(key)))
		for j := 0; j < 100; j++ {
			assert.True(t, isAlreadyLocked(ld.MSetNX(key)))
		}
		assert.True(t, isReleaseSuccess(ld.MDel(key)))
		assert.False(t, isReleaseSuccess(ld.MDel(key)))
	}
}
