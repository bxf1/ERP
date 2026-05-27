package builder

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConversationManager handles the lifecycle and state transitions of builder conversations.
type ConversationManager struct {
	mu             sync.RWMutex
	sessions       map[string]*Conversation
	store          ConversationStore
}

// ConversationStore is the persistence interface for conversations.
type ConversationStore interface {
	Save(c *Conversation) error
	Load(sessionID string) (*Conversation, error)
}

// NewConversationManager creates a new conversation manager.
func NewConversationManager(store ConversationStore) *ConversationManager {
	return &ConversationManager{
		sessions: make(map[string]*Conversation),
		store:    store,
	}
}

// StartSession creates a new builder conversation session.
func (cm *ConversationManager) StartSession() *Conversation {
	session := &Conversation{
		SessionID: uuid.New().String(),
		State:     StateRequirements,
		Messages:  make([]Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cm.mu.Lock()
	cm.sessions[session.SessionID] = session
	cm.mu.Unlock()
	return session
}

// GetSession retrieves an active conversation session.
func (cm *ConversationManager) GetSession(sessionID string) (*Conversation, error) {
	cm.mu.RLock()
	session, ok := cm.sessions[sessionID]
	cm.mu.RUnlock()
	if ok {
		return session, nil
	}
	if cm.store != nil {
		return cm.store.Load(sessionID)
	}
	return nil, fmt.Errorf("session %s not found", sessionID)
}

// AddMessage appends a message to the conversation and updates the timestamp.
func (cm *ConversationManager) AddMessage(sessionID, role, content string) error {
	session, err := cm.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.Messages = append(session.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	session.UpdatedAt = time.Now()
	if cm.store != nil {
		return cm.store.Save(session)
	}
	return nil
}

// TransitionState validates and executes a state transition.
// Valid transitions:
//
//	requirements → solution → confirmation → creation → done
//	confirmation → requirements (user wants changes)
//	confirmation → solution (tweak the proposal)
func (cm *ConversationManager) TransitionState(sessionID string, newState ConversationState) error {
	session, err := cm.GetSession(sessionID)
	if err != nil {
		return err
	}
	if !isValidTransition(session.State, newState) {
		return fmt.Errorf("invalid state transition: %s → %s", session.State, newState)
	}
	session.State = newState
	session.UpdatedAt = time.Now()
	if cm.store != nil {
		return cm.store.Save(session)
	}
	return nil
}

// SetProposedConfigs stores the generated model configs for the current session.
func (cm *ConversationManager) SetProposedConfigs(sessionID string, configs []ModelConfig) error {
	session, err := cm.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.ProposedConfigs = configs
	session.UpdatedAt = time.Now()
	if cm.store != nil {
		return cm.store.Save(session)
	}
	return nil
}

// RecordCreatedModel records a successfully created model in the session.
func (cm *ConversationManager) RecordCreatedModel(sessionID, modelID string) error {
	session, err := cm.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.CreatedModels = append(session.CreatedModels, modelID)
	session.UpdatedAt = time.Now()
	if cm.store != nil {
		return cm.store.Save(session)
	}
	return nil
}

func isValidTransition(from, to ConversationState) bool {
	switch from {
	case StateRequirements:
		return to == StateSolution
	case StateSolution:
		return to == StateConfirmation || to == StateRequirements
	case StateConfirmation:
		return to == StateCreation || to == StateRequirements || to == StateSolution
	case StateCreation:
		return to == StateDone || to == StateRequirements
	case StateDone:
		return to == StateRequirements
	default:
		return false
	}
}
