package paxosImpl

import (
	"sync"
)

// StateManager manages the state of the Paxos acceptor.
type StateManager struct {
	minProposal      int64        // The highest proposal number seen so far.
	acceptedProposal int64        // The proposal number that has been accepted.
	acceptedValue    interface{}  // The value associated with the accepted proposal.
	mu               sync.RWMutex // Mutex for thread-safe access to state variables.
}

// NewStateManager initializes and returns a new StateManager instance.
func NewStateManager() *StateManager {
	return &StateManager{
		minProposal:      0,
		acceptedProposal: 0,
		acceptedValue:    nil,
	}
}

// UpdateState allows updating any of the state variables
func (s *StateManager) UpdateState(minProposal *int64, acceptedProposal *int64, acceptedValue *interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if minProposal != nil {
		s.minProposal = *minProposal
	}
	if acceptedProposal != nil {
		s.acceptedProposal = *acceptedProposal
	}
	if acceptedValue != nil {
		s.acceptedValue = *acceptedValue
	}
}

// GetState returns the current state values in a thread-safe manner.
func (s *StateManager) GetState() (int64, int64, interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.minProposal, s.acceptedProposal, s.acceptedValue
}

// GetMinProposal returns the current minProposal value.
func (s *StateManager) GetMinProposal() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.minProposal
}

// GetAcceptedProposal returns the current acceptedProposal value.
func (s *StateManager) GetAcceptedProposal() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.acceptedProposal
}

// GetAcceptedValue returns the current acceptedValue.
func (s *StateManager) GetAcceptedValue() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.acceptedValue
}
