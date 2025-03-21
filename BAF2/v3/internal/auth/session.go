package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// SessionManager manages sessions between clients and devices
type SessionManager struct {
	sessionDir string
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionDir string) *SessionManager {
	// Use default directory if not provided
	if sessionDir == "" {
		sessionDir = "sessions"
	}
	
	// Create session directory if it doesn't exist
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		os.MkdirAll(sessionDir, 0755)
	}
	
	return &SessionManager{
		sessionDir: sessionDir,
	}
}

// SaveSession saves a session to a file
func (sm *SessionManager) SaveSession(session *Session) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}
	
	// Create filename
	filename := fmt.Sprintf("%s-%s-%s.json", session.ClientID, session.DeviceID, session.SessionID)
	sessionPath := filepath.Join(sm.sessionDir, filename)
	
	// Save session file
	if err := ioutil.WriteFile(sessionPath, sessionJSON, 0600); err != nil {
		return errors.Wrap(err, "failed to save session file")
	}
	
	return nil
}

// GetSession retrieves a session for a client and device
func (sm *SessionManager) GetSession(clientID, deviceID string) (*Session, error) {
	// Find matching session file
	pattern := filepath.Join(sm.sessionDir, fmt.Sprintf("%s-%s-*.json", clientID, deviceID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for session files")
	}
	
	if len(matches) == 0 {
		return nil, errors.Errorf("no active session found for client %s and device %s", clientID, deviceID)
	}
	
	// Use the first match (there should only be one active session per client-device pair)
	sessionPath := matches[0]
	
	// Read session file
	sessionJSON, err := ioutil.ReadFile(sessionPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read session file")
	}
	
	// Parse session
	var session Session
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return nil, errors.Wrap(err, "failed to parse session")
	}
	
	return &session, nil
}

// GetSessionByID retrieves a session by its ID
func (sm *SessionManager) GetSessionByID(sessionID string) (*Session, error) {
	// Find matching session file
	pattern := filepath.Join(sm.sessionDir, fmt.Sprintf("*-*-%s.json", sessionID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for session file")
	}
	
	if len(matches) == 0 {
		return nil, errors.Errorf("session %s not found", sessionID)
	}
	
	// Read session file
	sessionPath := matches[0]
	sessionJSON, err := ioutil.ReadFile(sessionPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read session file")
	}
	
	// Parse session
	var session Session
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return nil, errors.Wrap(err, "failed to parse session")
	}
	
	return &session, nil
}

// RemoveSession removes a session file
func (sm *SessionManager) RemoveSession(clientID, deviceID string) error {
	// Find matching session file
	pattern := filepath.Join(sm.sessionDir, fmt.Sprintf("%s-%s-*.json", clientID, deviceID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return errors.Wrap(err, "failed to search for session files")
	}
	
	if len(matches) == 0 {
		return errors.Errorf("no active session found for client %s and device %s", clientID, deviceID)
	}
	
	// Remove all matching files (should only be one)
	for _, sessionPath := range matches {
		if err := os.Remove(sessionPath); err != nil {
			return errors.Wrap(err, "failed to remove session file")
		}
	}
	
	return nil
}

// RemoveSessionByID removes a session file by its ID
func (sm *SessionManager) RemoveSessionByID(sessionID string) error {
	// Find matching session file
	pattern := filepath.Join(sm.sessionDir, fmt.Sprintf("*-*-%s.json", sessionID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return errors.Wrap(err, "failed to search for session file")
	}
	
	if len(matches) == 0 {
		return errors.Errorf("session %s not found", sessionID)
	}
	
	// Remove session file
	sessionPath := matches[0]
	if err := os.Remove(sessionPath); err != nil {
		return errors.Wrap(err, "failed to remove session file")
	}
	
	return nil
}

// ListActiveSessions lists all active sessions
func (sm *SessionManager) ListActiveSessions() ([]*Session, error) {
	// Find all session files
	pattern := filepath.Join(sm.sessionDir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for session files")
	}
	
	sessions := make([]*Session, 0, len(matches))
	
	// Parse each session file
	for _, sessionPath := range matches {
		sessionJSON, err := ioutil.ReadFile(sessionPath)
		if err != nil {
			log.Warnf("Failed to read session file %s: %v", sessionPath, err)
			continue
		}
		
		var session Session
		if err := json.Unmarshal(sessionJSON, &session); err != nil {
			log.Warnf("Failed to parse session file %s: %v", sessionPath, err)
			continue
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

// GetActiveSessionsForClient lists all active sessions for a client
func (sm *SessionManager) GetActiveSessionsForClient(clientID string) ([]*Session, error) {
	// Find matching session files
	pattern := filepath.Join(sm.sessionDir, fmt.Sprintf("%s-*.json", clientID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for session files")
	}
	
	sessions := make([]*Session, 0, len(matches))
	
	// Parse each session file
	for _, sessionPath := range matches {
		sessionJSON, err := ioutil.ReadFile(sessionPath)
		if err != nil {
			log.Warnf("Failed to read session file %s: %v", sessionPath, err)
			continue
		}
		
		var session Session
		if err := json.Unmarshal(sessionJSON, &session); err != nil {
			log.Warnf("Failed to parse session file %s: %v", sessionPath, err)
			continue
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}
