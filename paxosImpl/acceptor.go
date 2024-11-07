package paxosImpl

import (
	"paxos/communication"
)

type Acceptor struct {
	id                int64                          // Unique ID for the Acceptor
	stateManager      *StateManager                  // StateManager to manage shared state
	tcpCommunicator   *communication.TcpCommunicator // Communicator to send and receive messages
	prepareMessagesCh chan communication.Message     // Channel to receive Prepare messages
	acceptMessagesCh  chan communication.Message     // Channel to receive Accept messages
}

// NewAcceptor initializes a new Acceptor instance.
func NewAcceptor(id int64, prepareMessagesCh chan communication.Message,
	acceptMessagesCh chan communication.Message, tcpCommunicator *communication.TcpCommunicator, stateManager *StateManager) *Acceptor {
	return &Acceptor{
		id:                id,
		stateManager:      stateManager,
		tcpCommunicator:   tcpCommunicator,
		prepareMessagesCh: prepareMessagesCh,
		acceptMessagesCh:  acceptMessagesCh,
	}
}

func (a *Acceptor) Listen() {
	for {
		select {
		case message := <-a.prepareMessagesCh:
			a.handlePrepareMessage(message)
		case message := <-a.acceptMessagesCh:
			a.handleAcceptMessage(message)
		}
	}
}

// handlePrepareMessage processes a Prepare message.
func (a *Acceptor) handlePrepareMessage(message communication.Message) {
	// Check if the proposal is greater than the current min proposal
	if message.Payload.Proposal > a.stateManager.GetMinProposal() {
		a.stateManager.UpdateState(&message.Payload.Proposal, nil, nil)
	}
	err := a.tcpCommunicator.SendPromiseMessage(message.Header.SenderID, a.stateManager.GetAcceptedProposal(), a.stateManager.GetAcceptedValue())
	if err != nil {
		panic(err.Error())
	}

}

// handleAcceptMessage processes an Accept message.
func (a *Acceptor) handleAcceptMessage(message communication.Message) {
	// Check if the proposal is greater than or equal to the current min proposal
	if message.Payload.Proposal >= a.stateManager.GetMinProposal() {
		a.stateManager.UpdateState(&message.Payload.Proposal, &message.Payload.Proposal, &message.Payload.Value)
	}
	minProposal := a.stateManager.GetMinProposal()
	err := a.tcpCommunicator.SendAcceptedMessage(message.Header.SenderID, minProposal, nil)
	if err != nil {
		panic(err.Error())
	}
}
