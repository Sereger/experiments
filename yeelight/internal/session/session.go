package session

import (
	"sync"
	"time"

	"github.com/Sereger/experiments/yeelight/internal/yeelight"
)

type (
	Session struct {
		LastTargets  map[*yeelight.Yeelight]struct{}
		Devices      []*yeelight.Yeelight
		DeviceTokens map[string][]*yeelight.Yeelight
		touch        time.Time
	}

	Storage struct {
		mu   sync.RWMutex
		data map[string]*Session
	}
)

func NewStorage() *Storage {
	s := &Storage{
		data: make(map[string]*Session),
	}

	go func() {
		for {
			time.Sleep(time.Minute)
			s.gc()
		}
	}()

	return s
}

func (s *Storage) ResolveSession(key string) *Session {
	s.mu.RLock()
	sess, ok := s.data[key]
	s.mu.RUnlock()
	if ok {
		sess.touch = time.Now()
		return sess
	}

	newSess := &Session{
		touch: time.Now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = newSess

	return newSess
}

func (s *Storage) gc() {
	s.mu.Lock()
	defer s.mu.Unlock()

	moment := time.Now().Add(-3 * time.Minute)
	for key, sess := range s.data {
		if sess.touch.Before(moment) {
			delete(s.data, key)
		}
	}
}
