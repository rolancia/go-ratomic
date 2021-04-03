package ratomic

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewLocalDriver(t *testing.T) {
	ld := NewLocalDriver("lock", 3*time.Millisecond)

	key := "tae"

	for i := 0; i < 10; i++ {
		num, _ := ld.MSetNX(key)
		assert.False(t, isAlreadyLocked(num))
		for j := 0; j < 100; j++ {
			num, _ = ld.MSetNX(key)
			assert.True(t, isAlreadyLocked(num))
		}
		num, _ = ld.Del(key)
		assert.True(t, isReleaseSuccess(num))
		num, _ = ld.Del(key)
		assert.False(t, isReleaseSuccess(num))
	}
}
