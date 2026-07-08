package domain

import (
	"fmt"
	"time"
)

// Role indica quién envió el mensaje.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message representa un mensaje en la conversación de una sesión.
type Message struct {
	Role    Role
	Content string
	Time    time.Time
}

// Session agrupa la configuración de DB, el DDL y el historial de mensajes.
type Session struct {
	ID       string
	Name     string
	Config   ConnectionConfig
	DDLInfo  DDLInfo
	Messages []Message
	Created  time.Time
}

// SessionManager gestiona el mapa de sesiones y la sesión activa.
type SessionManager struct {
	Sessions map[string]*Session
	ActiveID string
}

// NewSessionManager inicializa el gestor de sesiones.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*Session),
	}
}

// AddSession registra una nueva sesión y la activa automáticamente.
func (sm *SessionManager) AddSession(s *Session) {
	sm.Sessions[s.ID] = s
	sm.ActiveID = s.ID
}

// SwitchSession cambia la sesión activa por su ID.
func (sm *SessionManager) SwitchSession(id string) bool {
	if _, ok := sm.Sessions[id]; !ok {
		return false
	}
	sm.ActiveID = id
	return true
}

// GetActive devuelve la sesión actualmente activa o nil.
func (sm *SessionManager) GetActive() *Session {
	if sm.ActiveID == "" {
		return nil
	}
	return sm.Sessions[sm.ActiveID]
}

// NextID genera un ID simple incremental basado en la cantidad de sesiones.
func (sm *SessionManager) NextID() string {
	return fmt.Sprintf("%d", len(sm.Sessions)+1)
}
