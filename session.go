package main

import (
	"container/list"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GehirnInc/crypt/sha512_crypt"
)

type adminUser struct {
	id       any
	username any
	level    int64
	expires  time.Time
}

type sessionEntry struct {
	token string
	user  adminUser
}

type sessions struct {
	mu       sync.Mutex
	capacity int
	entries  map[string]*list.Element
	lru      *list.List
}

func newSessions(capacity int) *sessions {
	return &sessions{capacity: capacity, entries: make(map[string]*list.Element), lru: list.New()}
}

func (s *sessions) put(token string, user adminUser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.entries[token]; e != nil {
		e.Value = sessionEntry{token, user}
		s.lru.MoveToFront(e)
		return
	}
	s.entries[token] = s.lru.PushFront(sessionEntry{token, user})
	if s.lru.Len() > s.capacity {
		e := s.lru.Back()
		delete(s.entries, e.Value.(sessionEntry).token)
		s.lru.Remove(e)
	}
}

func (s *sessions) get(token string) (adminUser, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entries[token]
	if e == nil {
		return adminUser{}, false
	}
	u := e.Value.(sessionEntry).user
	if u.expires.Before(time.Now()) {
		delete(s.entries, token)
		s.lru.Remove(e)
		return adminUser{}, false
	}
	s.lru.MoveToFront(e)
	return u, true
}

func (s *sessions) remove(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.entries[token]; e != nil {
		delete(s.entries, token)
		s.lru.Remove(e)
	}
}

func adminPassword(password, salt string) string {
	h := sha512.Sum512([]byte(password + salt))
	return hex.EncodeToString(h[:])[:64]
}

func mailboxPassword(password string) (string, error) {
	var salt [8]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", err
	}
	// Same $6$ SHA512-crypt, 16-character salt and default 5000 rounds as crypt.crypt.
	return sha512_crypt.New().Generate([]byte(password), []byte("$6$"+hex.EncodeToString(salt[:])))
}

func generateToken() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	var token strings.Builder
	// Python's hex(ord(byte))[2:] deliberately did not zero-pad bytes below 16.
	for _, b := range data {
		token.WriteString(strconv.FormatUint(uint64(b), 16))
	}
	return token.String(), nil
}
