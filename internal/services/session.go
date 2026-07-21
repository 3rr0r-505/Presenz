// internal/services/session.go

package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/3rr0r-505/Presenz/internal/config"
)

type Session struct {
	mu           sync.Mutex
	active       bool
	sessionID    string
	sessionCode  string
	tableName    string
	maxCount     int
	currentCount int
	dbPath       string
}

// Start initializes a new attendance session. Not safe for concurrent
// calls with itself (mirrors Python: only called once at boot, before
// any HTTP traffic exists).
func (s *Session) Start(cfg *config.Config, maxCount int, course, batch, dbFilename string) error {
	if s.active {
		return errors.New("session already active")
	}

	sessionID, err := generateToken(cfg.Session.SessionIDLength)
	if err != nil {
		return fmt.Errorf("generating session id: %w", err)
	}
	sessionCode, err := generateToken(cfg.Session.SessionCodeLength)
	if err != nil {
		return fmt.Errorf("generating session code: %w", err)
	}

	now := time.Now().Format("02-01-06-1504")
	safeCourse := sanitizeIdentifier(course)
	safeBatch := sanitizeIdentifier(batch)

	if strings.Contains(dbFilename, "/") || strings.Contains(dbFilename, "..") {
		return errors.New("invalid database filename")
	}

	s.sessionID = sessionID
	s.sessionCode = sessionCode
	s.tableName = fmt.Sprintf("%s-%s-%s", now, safeCourse, safeBatch)
	s.maxCount = maxCount
	s.currentCount = 0
	s.dbPath = cfg.Database.BasePath + dbFilename
	s.active = true

	return nil
}

// TryAcceptSubmission atomically checks capacity and reserves a slot.
func (s *Session) TryAcceptSubmission() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active && s.currentCount < s.maxCount {
		s.currentCount++
		return true
	}
	return false
}

// DecrementCount releases a reserved slot (e.g. on duplicate-roll insert failure).
func (s *Session) DecrementCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentCount--
}

func (s *Session) IsFull() bool {
	return s.currentCount >= s.maxCount
}

func (s *Session) ValidateSessionCode(code string) bool {
	return s.active && code == s.sessionCode
}

func (s *Session) GetSessionCode() string { return s.sessionCode }
func (s *Session) GetTableName() string   { return s.tableName }
func (s *Session) DBPath() string         { return s.dbPath }
func (s *Session) Active() bool           { return s.active }

func (s *Session) End() {
	s.active = false
	s.sessionID = ""
	s.sessionCode = ""
	s.tableName = ""
	s.currentCount = 0
	s.dbPath = ""
}

// generateToken produces a cryptographically random alphanumeric string
// (uppercase A-Z + 0-9), matching Python's secrets.choice usage.
func generateToken(length int) (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}

// sanitizeIdentifier uppercases and strips non-alphanumeric characters,
// used to build safe table name components from course/batch input.
func sanitizeIdentifier(value string) string {
	var sb strings.Builder
	for _, r := range strings.ToUpper(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
