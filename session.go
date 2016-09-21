package main

import "sync"

type SessionStorage struct {
	*sync.RWMutex
	data map[string]string
}

func NewSessionStorage() *SessionStorage {
	return &SessionStorage{&sync.RWMutex{}, make(map[string]string)}
}

func (s *SessionStorage) Get(apiKey string) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	id, ok := s.data[apiKey]
	return id, ok
}

func (s *SessionStorage) Set(apiKey, userId string) {
	s.Lock()
	defer s.Unlock()
	s.data[apiKey] = userId
}
