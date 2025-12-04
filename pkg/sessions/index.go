package sessions

import (
	"sync"
)

type device_id int

type Session struct {
	TrainingID int
	Rep        int
	MovementID int
	DeviceID   int
}

type SessionManager struct {
	mu sync.RWMutex
	s  map[device_id]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		s: make(map[device_id]*Session),
	}
}

func (m *SessionManager) Set(id int, sess *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.s[device_id(id)] = sess
}

func (m *SessionManager) Get(id int) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ex := m.s[device_id(id)]
	return val, ex
}

func (m *SessionManager) Update(id int, fn func(s *Session)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.s[device_id(id)]
	if !ok {
		return
	}

	fn(sess)
}

func (m *SessionManager) Delete(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.s, device_id(id))
}

func (m *SessionManager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*Session, 0, len(m.s))
	for _, v := range m.s {
		out = append(out, v)
	}

	return out
}
