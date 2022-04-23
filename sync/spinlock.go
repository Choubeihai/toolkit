package sync

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	value int32
}

func (lock *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&lock.value, 0, 1) {
		runtime.Gosched()
	}
}

func (lock *SpinLock) Unlock() {
	atomic.StoreInt32(&lock.value, 0)
}
