package paxosImpl

import (
	"fmt"
	"paxos/communication"
	"sync"
)

// Proposer represents a Paxos proposer that initiates the Prepare and Accept phases.
type Proposer struct {
	id                 int64                          // Unique ID for the Proposer
	proposalNumber     int64                          // The proposal number for this instance
	minProposalNumber  int64                          // The minimum proposal number seen so far
	value              interface{}                    // The value the Proposer wants to propose
	acceptedValue      interface{}                    // The value that has been accepted
	quorum             []int64                        // The set of acceptor IDs to communicate with
	promiseResponses   int64                          // Tracks the responses from Acceptors
	acceptResponses    int64                          // Tracks the responses from Acceptors
	stateManager       *StateManager                  // StateManager to manage shared state
	tcpCommunicator    *communication.TcpCommunicator // Communicator to send and receive messages
	promiseMessagesCh  chan communication.Message     // Channel to receive Promise messages
	acceptedMessagesCh chan communication.Message     // Channel to receive Accepted messages
	sendProposalCh     chan bool                      // Channel to send proposal
	mu                 sync.Mutex                     // Mutex for thread-safe updates to responses
}

// NewProposer initializes a new Proposer instance.
func NewProposer(id int64, proposalNumber int64, value interface{}, quorum []int64, promiseMessagesCh chan communication.Message,
	acceptedMessagesCh chan communication.Message, sendProposalCh chan bool, tcpCommunicator *communication.TcpCommunicator, stateManager *StateManager) *Proposer {
	return &Proposer{
		id:                 int64(id),
		proposalNumber:     int64(proposalNumber),
		value:              value,
		quorum:             quorum,
		promiseResponses:   0,
		acceptResponses:    0,
		stateManager:       stateManager,
		tcpCommunicator:    tcpCommunicator,
		promiseMessagesCh:  promiseMessagesCh,
		acceptedMessagesCh: acceptedMessagesCh,
		sendProposalCh:     sendProposalCh,
		minProposalNumber:  0,
		acceptedValue:      nil,
	}
}

// Prepare initiates the Prepare phase by sending Prepare messages to the quorum.
func (p *Proposer) sendProposal() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.minProposalNumber > p.proposalNumber {
		p.proposalNumber = p.minProposalNumber
	}

	if p.acceptedValue != nil {
		p.value = p.acceptedValue
		p.acceptedValue = nil
	}

	p.promiseResponses = 0
	p.proposalNumber += 1 // Increment proposal number

	// Send Prepare message to each acceptor in the quorum
	p.stateManager.UpdateState(&p.proposalNumber, nil, nil)

	for _, acceptorID := range p.quorum {
		if acceptorID == p.id {
			p.promiseResponses += 1
		} else {
			err := p.tcpCommunicator.SendPrepareMessage(acceptorID, p.proposalNumber, p.value)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

func (p *Proposer) Listen() {
	for {
		select {
		case <-p.sendProposalCh:
			p.sendProposal()
		case message := <-p.promiseMessagesCh:
			p.handlePromiseMessage(message)
		case message := <-p.acceptedMessagesCh:
			p.handleAcceptedMessage(message)
		}
	}
}

// handlePromiseMessage processes a Promise message.
func (p *Proposer) handlePromiseMessage(message communication.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.promiseResponses += 1
	if message.Payload.Proposal > p.minProposalNumber {
		p.minProposalNumber = message.Payload.Proposal
	}
	if message.Payload.Value != nil {
		p.acceptedValue = message.Payload.Value
	}
	p.acceptResponses = 0
	if p.promiseResponses == int64(len(p.quorum)) {
		// Ready to send accept message
		if p.acceptedValue != nil {
			p.value = p.acceptedValue
		}
		p.stateManager.UpdateState(&p.proposalNumber, &p.proposalNumber, &p.value)
		for _, acceptorID := range p.quorum {
			if acceptorID == p.id {
				p.acceptResponses += 1
			} else {
				err := p.tcpCommunicator.SendAcceptMessage(acceptorID, p.proposalNumber, p.value)
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}

}

// handleAcceptedMessage processes an Accepted message.
func (p *Proposer) handleAcceptedMessage(message communication.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.acceptResponses += 1

	if message.Payload.Proposal > p.minProposalNumber {
		p.minProposalNumber = message.Payload.Proposal
	}
	if message.Payload.Value != p.acceptedValue {
		p.acceptedValue = message.Payload.Value // helps for the next prepare message incase this is not the final value
	}
	if p.acceptResponses == int64(len(p.quorum)) {
		if p.minProposalNumber <= p.proposalNumber {
			// Choose the value
			fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", p.id, "chose", "chose", p.stateManager.GetAcceptedValue(), message.Payload.Proposal)
		} else {
			// Send proposal again
			p.sendProposalCh <- true
		}
	}
}
