package contexted_mutex

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	ctn, cnlFn := context.WithCancel(context.Background())
	l := NewLock()
	err := l.Lock(ctn)
	if err != nil {
		t.Fatal("lock should be taken")
		t.FailNow()
	}

	tik := time.NewTimer(time.Second)
	defer tik.Stop()

	go func() {
		<-tik.C
		cnlFn()
	}()

	s := time.Now()
	err = l.Lock(ctn)
	d := time.Since(s)
	if err == nil {
		t.Fatal("expected error here")
	}
	t.Logf("collect err: [%s]", err)

	t.Logf("lock wait time: [%s], expect: ~1s", d)
	l.Unlock()
}

func TestRLock(t *testing.T) {
	ctn, cnlFn := context.WithCancel(context.Background())
	l := NewLock()
	err := l.RLock(ctn)
	if err != nil {
		t.Fatal("rlock should be taken")
		t.FailNow()
	}

	err = l.RLock(ctn)
	if err != nil {
		t.Fatal("rlock should be taken")
		t.FailNow()
	}

	time.AfterFunc(time.Second, func() {
		cnlFn()
	})
	s := time.Now()
	err = l.Lock(ctn)
	d := time.Since(s)
	if err == nil {
		t.Fatal("expected error here")
	}
	t.Logf("collect err: [%s]", err)

	t.Logf("lock wait time: [%s], expect: ~1s", d)
	l.Unlock()
}

func TestRLock2(t *testing.T) {
	ctn := context.Background()
	l := NewLock()
	err := l.RLock(ctn)
	if err != nil {
		t.Fatal("rlock should be taken")
		t.FailNow()
	}

	err = l.RLock(ctn)
	if err != nil {
		t.Fatal("rlock should be taken")
		t.FailNow()
	}

	time.AfterFunc(time.Second, func() {
		l.RUnlock()
		l.RUnlock()
	})
	s := time.Now()
	err = l.Lock(ctn)
	d := time.Since(s)
	if err != nil {
		t.Fatal("lock should be taken")
	}

	t.Logf("lock wait time: [%s], expect: ~1s", d)
	l.Unlock()
}

// This test for comparing waiting results
func TestMutex(t *testing.T) {
	l := sync.RWMutex{}
	l.RLock()
	l.RLock()

	time.AfterFunc(time.Second, func() {
		l.RUnlock()
		l.RUnlock()
	})
	s := time.Now()
	l.Lock()
	d := time.Since(s)

	t.Logf("lock wait time: [%s], expect: ~1s", d)
	l.Unlock()
}
