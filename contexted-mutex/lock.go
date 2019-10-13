package contexted_mutex

import (
	"context"
	"sync/atomic"
)

const (
	unlocked = uint32(0)
	locked   = uint32(1)
)

type (
	RWLock struct {
		addrLock *uint32
		addrCnr  *uint32
	}
)

func NewLock() *RWLock {
	l := &RWLock{
		addrLock: new(uint32),
		addrCnr:  new(uint32),
	}

	return l
}

func (l *RWLock) Lock(ctn context.Context) error {

lockLoop:
	for {
		select {
		case <-ctn.Done():
			return ctn.Err()
		default:
			if atomic.CompareAndSwapUint32(l.addrLock, unlocked, locked) {
				break lockLoop
			}
		}
	}

RlockLoop:
	for {
		select {
		case <-ctn.Done():
			l.Unlock()
			return ctn.Err()
		default:
			if atomic.LoadUint32(l.addrCnr) == 0 {
				break RlockLoop
			}
		}
	}

	return nil
}

func (l *RWLock) RLock(ctn context.Context) error {
lockLoop:
	for {
		select {
		case <-ctn.Done():
			return ctn.Err()
		default:
			if atomic.CompareAndSwapUint32(l.addrLock, unlocked, locked) {
				break lockLoop
			}
		}
	}

	l.rIncr()
	l.Unlock()
	return nil
}

func (l *RWLock) locked() bool {
	return atomic.LoadUint32(l.addrLock) == locked
}
func (l *RWLock) RUnlock() {
	l.rDecr()
}

func (l *RWLock) rIncr() uint32 {
	return atomic.AddUint32(l.addrCnr, 1)
}

func (l *RWLock) rDecr() uint32 {
	return atomic.AddUint32(l.addrCnr, ^uint32(0))
}

func (l *RWLock) Unlock() {
	atomic.StoreUint32(l.addrLock, 0)
}
