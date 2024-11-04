package main

import (
	"fmt"
	"os"
	"paxos/communication"
	"paxos/paxosImpl"
	"paxos/util"
	"time"
)

func main() {
	hostfile, proposerValue, waitTime := util.ParseFlags()
	hostRoles, quorumMap := util.ReadHostfile(hostfile)
	me, _ := os.Hostname()
	communicator := communication.NewTcpCommunicator()
	stateManager := paxosImpl.NewStateManager()

	connectionsEstablishedCh := make(chan bool)
	sendProposalCh := make(chan bool)
	incomingMessagesCh := make(chan communication.Message)
	prepareMessagesCh := make(chan communication.Message)
	acceptMessagesCh := make(chan communication.Message)
	promiseMessagesCh := make(chan communication.Message)
	acceptedMessagesCh := make(chan communication.Message)
	go communicator.Listen(incomingMessagesCh)

	for id, info := range hostRoles {
		if info.Hostname == me {
			communicator.SetSelfId(id)
			if len(info.Proposer) > 0 {
				for _, val := range info.Proposer {
					// Initiate the proposer
					proposer := paxosImpl.NewProposer(id, val, proposerValue, quorumMap[val], promiseMessagesCh, acceptedMessagesCh, sendProposalCh, communicator, stateManager)
					go proposer.Listen()
				}
			}
			// Initiate the acceptor
			acceptor := paxosImpl.NewAcceptor(id, prepareMessagesCh, acceptMessagesCh, communicator, stateManager)
			go acceptor.Listen()
		} else {
			communicator.AddPeer(id, info.Hostname)
		}
	}

	// Establish connections before moving foward
	go communicator.EstablishConnections(connectionsEstablishedCh)
	<-connectionsEstablishedCh

	// Go through the proposers through channel and start the proposal
	go func() {
		time.Sleep(time.Duration(waitTime) * time.Second)
		sendProposalCh <- true
	}()

	for message := range incomingMessagesCh {
		var messageType string
		switch message.Header.MessageType {
		case communication.PREPARE:
			messageType = "prepare"
			prepareMessagesCh <- message
		case communication.PROMISE:
			messageType = "prepare_ack"
			promiseMessagesCh <- message
		case communication.ACCEPT:
			messageType = "accept"
			acceptMessagesCh <- message
		case communication.ACCEPTED:
			messageType = "accept_ack"
			acceptedMessagesCh <- message
		}
		fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", message.Header.SenderID, "received", messageType, message.Payload.Value, message.Payload.Proposal)

	}
}
