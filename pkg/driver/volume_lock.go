package driver

import (
	"sync"
)

type volumeLock struct {
	cond   sync.Cond
	locked map[string]struct{}
}

func newVolumeLock() *volumeLock {
	return &volumeLock{
		cond:   *sync.NewCond(&sync.Mutex{}),
		locked: map[string]struct{}{},
	}
}

func (l *volumeLock) LockVolume(volume string) func() {
	l.cond.L.Lock()
	defer l.cond.L.Unlock()

	for {
		if _, locked := l.locked[volume]; !locked {
			break
		}

		l.cond.Wait()
	}

	l.locked[volume] = struct{}{}

	return func() {
		l.cond.L.Lock()
		defer l.cond.L.Unlock()

		delete(l.locked, volume)
		l.cond.Broadcast()
	}
}

func (l *volumeLock) LockVolumeWithSnapshot(volume string, snapshot string) func() {
	unlockVol := l.LockVolume(volume)
	unlockSnap := l.LockVolume(snapshot)
	return func() {
		unlockVol()
		unlockSnap()
	}
}
