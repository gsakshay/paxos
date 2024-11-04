package communication

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const TCPPort = "8888"

type TcpCommunicator struct {
	selfId      int64              // The ID of the current peer.
	peers       map[int64]string   // Maps peer IDs to their hostnames.
	connections map[int64]net.Conn // Maps peer IDs to their active TCP connections.
	mu          sync.Mutex         // Mutex for thread-safe access to connections.
}

func NewTcpCommunicator() *TcpCommunicator {
	return &TcpCommunicator{
		selfId:      0,
		peers:       make(map[int64]string),
		connections: make(map[int64]net.Conn),
	}
}

// Sets the ID of the current peer.
func (c *TcpCommunicator) SetSelfId(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selfId = id
}

// Adds a peer to the list of peers.
func (c *TcpCommunicator) AddPeer(id int64, hostname string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.peers[id] = hostname
}

// Tries to establish TCP connections to all peers until successful.
func (c *TcpCommunicator) EstablishConnections(connectedCh chan bool) {
	var wg sync.WaitGroup

	for id, name := range c.peers {
		wg.Add(1)
		go func(id int64, name string) {
			defer wg.Done()
			for {
				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", name, TCPPort))
				if err == nil {
					c.mu.Lock()
					c.connections[id] = conn
					c.mu.Unlock()
					break
				} else {
					time.Sleep(500 * time.Millisecond)
				}

			}
		}(id, name)
	}

	wg.Wait()
	connectedCh <- true // Signal that all connections are established
}

func (c *TcpCommunicator) sendMessage(id int64, message []byte) error { // int64ended to be private
	c.mu.Lock()
	conn, exists := c.connections[id]
	c.mu.Unlock()

	if !exists {
		log.Fatalf("no connection found for peer %v", id)
	}

	_, err := conn.Write(message)
	return err
}

func (c *TcpCommunicator) Listen(messageCh chan Message) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", TCPPort))
	if err != nil {
		log.Fatalf("Failed to start listener on port %s: %v\n", TCPPort, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			for {
				// Read and parse the message
				fullMessage, err := readAndParseMessage(conn)
				if err != nil {
					fmt.Printf("Failed to read and parse message: %v\n", err)
					break
				}
				// Handle the message
				messageCh <- fullMessage
			}
		}(conn)
	}
}

func (c *TcpCommunicator) sendPaxosMessage(targetId int64, messageType int64, proposalNumber int64, value interface{}) error {
	message := Message{
		Header: MessageHeader{
			SenderID:    c.selfId,
			MessageType: messageType,
			PayloadSize: 0, // Will be calculated in ConvertToBinary
		},
		Payload: PaxosMessage{
			Proposal: proposalNumber,
			Value:    value,
		},
	}
	buf, err := ConvertToBinary(message)
	if err != nil {
		return fmt.Errorf("failed to convert message to binary: %v", err)
	}
	return c.sendMessage(targetId, buf)
}

func (c *TcpCommunicator) SendPrepareMessage(targetId int64, proposalNumber int64, value interface{}) error {
	err := c.sendPaxosMessage(targetId, PREPARE, proposalNumber, value)
	if err == nil {
		fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", c.selfId, "sent", "prepare", value, proposalNumber)
	}
	return err
}

func (c *TcpCommunicator) SendAcceptMessage(targetId int64, proposalNumber int64, value interface{}) error {
	err := c.sendPaxosMessage(targetId, ACCEPT, proposalNumber, value)
	if err == nil {
		fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", c.selfId, "sent", "accept", value, proposalNumber)
	}
	return err
}

func (c *TcpCommunicator) SendPromiseMessage(targetId int64, acceptedProposal int64, acceptedValue interface{}) error {
	err := c.sendPaxosMessage(targetId, PROMISE, acceptedProposal, acceptedValue)
	if err == nil {
		fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", c.selfId, "sent", "prepare_ack", acceptedValue, acceptedProposal)
	}
	return err
}

func (c *TcpCommunicator) SendAcceptedMessage(targetId int64, acceptedProposal int64, acceptedValue interface{}) error {
	err := c.sendPaxosMessage(targetId, ACCEPTED, acceptedProposal, acceptedValue)
	if err == nil {
		fmt.Printf("{\"peer_id\": %v, \"action\": \"%v\", \"message_type\":\"%v\", \"message_value\":\"%v\", \"proposal_num\": %v}\n", c.selfId, "sent", "accept_ack", acceptedValue, acceptedProposal)
	}
	return err
}
