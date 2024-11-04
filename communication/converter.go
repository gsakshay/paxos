package communication

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
)

const (
	// Message Types
	PREPARE  = 1
	PROMISE  = 2
	ACCEPT   = 3
	ACCEPTED = 4
	// DataTypes
	NIL     = 0
	INT64   = 1
	FLOAT64 = 2
	STRING  = 3
)

type MessageHeader struct {
	SenderID    int64
	MessageType int64
	PayloadSize int64
}

type PaxosMessage struct {
	Proposal int64
	Value    interface{}
}

type Message struct {
	Header  MessageHeader
	Payload PaxosMessage
}

// ConvertToBinary converts a Message struct to a binary format using BigEndian.
func ConvertToBinary(message Message) ([]byte, error) {
	payloadBuf := new(bytes.Buffer)

	// Write the Proposal to the payload buffer
	if err := binary.Write(payloadBuf, binary.BigEndian, message.Payload.Proposal); err != nil {
		return nil, fmt.Errorf("failed to write Proposal: %v", err)
	}

	// Check if Value is nil
	if message.Payload.Value == nil {
		// Write the type indicator for nil
		if err := binary.Write(payloadBuf, binary.BigEndian, int64(NIL)); err != nil {
			return nil, fmt.Errorf("failed to write value type indicator: %v", err)
		}
		// No need to write any value data for nil
	} else {
		// Serialize the Value field with a type indicator
		switch v := message.Payload.Value.(type) {
		case int64:
			if err := binary.Write(payloadBuf, binary.BigEndian, int64(INT64)); err != nil {
				return nil, fmt.Errorf("failed to write value type indicator: %v", err)
			}
			if err := binary.Write(payloadBuf, binary.BigEndian, v); err != nil {
				return nil, fmt.Errorf("failed to write int64 Value: %v", err)
			}
		case float64:
			if err := binary.Write(payloadBuf, binary.BigEndian, int64(FLOAT64)); err != nil {
				return nil, fmt.Errorf("failed to write value type indicator: %v", err)
			}
			if err := binary.Write(payloadBuf, binary.BigEndian, v); err != nil {
				return nil, fmt.Errorf("failed to write float64 Value: %v", err)
			}
		case string:
			strBytes := []byte(v)
			if err := binary.Write(payloadBuf, binary.BigEndian, int64(STRING)); err != nil {
				return nil, fmt.Errorf("failed to write value type indicator: %v", err)
			}
			if err := binary.Write(payloadBuf, binary.BigEndian, int64(len(strBytes))); err != nil {
				return nil, fmt.Errorf("failed to write string length: %v", err)
			}
			if _, err := payloadBuf.Write(strBytes); err != nil {
				return nil, fmt.Errorf("failed to write string Value: %v", err)
			}
		default:
			return nil, fmt.Errorf("unsupported Value type: %v", reflect.TypeOf(v))
		}
	}

	// Compute PayloadSize
	message.Header.PayloadSize = int64(payloadBuf.Len())

	// Write the Header to the buffer
	headerBuf := new(bytes.Buffer)
	if err := binary.Write(headerBuf, binary.BigEndian, message.Header.SenderID); err != nil {
		return nil, fmt.Errorf("failed to write SenderID: %v", err)
	}
	if err := binary.Write(headerBuf, binary.BigEndian, message.Header.MessageType); err != nil {
		return nil, fmt.Errorf("failed to write MessageType: %v", err)
	}
	if err := binary.Write(headerBuf, binary.BigEndian, message.Header.PayloadSize); err != nil {
		return nil, fmt.Errorf("failed to write PayloadSize: %v", err)
	}

	// Combine header and payload
	buf := new(bytes.Buffer)
	buf.Write(headerBuf.Bytes())
	buf.Write(payloadBuf.Bytes())

	return buf.Bytes(), nil
}

func readFully(conn net.Conn, buffer []byte) error {
	totalRead := 0
	for totalRead < len(buffer) {
		n, err := conn.Read(buffer[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	return nil
}

func readAndParseMessage(conn net.Conn) (Message, error) {
	// Step 1: Read the header (12 bytes)
	header := make([]byte, 24)
	if err := readFully(conn, header); err != nil {
		return Message{}, fmt.Errorf("failed to read header: %v", err)
	}

	// Step 2: Read the header data int64o MessageHeader struct
	var msgHeader MessageHeader
	headerBuf := bytes.NewReader(header)
	if err := binary.Read(headerBuf, binary.BigEndian, &msgHeader.SenderID); err != nil {
		return Message{}, fmt.Errorf("failed to read SenderID: %v", err)
	}
	if err := binary.Read(headerBuf, binary.BigEndian, &msgHeader.MessageType); err != nil {
		return Message{}, fmt.Errorf("failed to read MessageType: %v", err)
	}
	if err := binary.Read(headerBuf, binary.BigEndian, &msgHeader.PayloadSize); err != nil {
		return Message{}, fmt.Errorf("failed to read PayloadSize: %v", err)
	}

	// Step 3: Read the payload
	payload := make([]byte, msgHeader.PayloadSize)
	if msgHeader.PayloadSize > 0 {
		if err := readFully(conn, payload); err != nil {
			return Message{}, fmt.Errorf("failed to read payload: %v", err)
		}
	}

	// Step 4: Combine header and payload, then convert from binary
	fullData := append(header, payload...)
	fullMessage, err := ConvertFromBinary(fullData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to convert from binary: %v", err)
	}

	return fullMessage, nil
}

// ConvertFromBinary converts binary data back int64o a Message struct.
func ConvertFromBinary(data []byte) (Message, error) {
	buf := bytes.NewReader(data)
	var header MessageHeader

	// Read the Header from the buffer
	if err := binary.Read(buf, binary.BigEndian, &header.SenderID); err != nil {
		return Message{}, fmt.Errorf("failed to read SenderID: %v", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &header.MessageType); err != nil {
		return Message{}, fmt.Errorf("failed to read MessageType: %v", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &header.PayloadSize); err != nil {
		return Message{}, fmt.Errorf("failed to read PayloadSize: %v", err)
	}

	// Read the Payload from the buffer
	var payload PaxosMessage
	if header.PayloadSize > 0 {
		if err := binary.Read(buf, binary.BigEndian, &payload.Proposal); err != nil {
			return Message{}, fmt.Errorf("failed to read Proposal: %v", err)
		}

		// Read the type indicator
		var valueType int64
		if err := binary.Read(buf, binary.BigEndian, &valueType); err != nil {
			return Message{}, fmt.Errorf("failed to read value type: %v", err)
		}

		switch valueType {
		case NIL:
			// Set Value to nil
			payload.Value = nil
		case INT64:
			var int64Value int64
			if err := binary.Read(buf, binary.BigEndian, &int64Value); err != nil {
				return Message{}, fmt.Errorf("failed to read int64 Value: %v", err)
			}
			payload.Value = int64Value
		case FLOAT64:
			var floatValue float64
			if err := binary.Read(buf, binary.BigEndian, &floatValue); err != nil {
				return Message{}, fmt.Errorf("failed to read float64 Value: %v", err)
			}
			payload.Value = floatValue
		case STRING:
			var strLen int64
			if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
				return Message{}, fmt.Errorf("failed to read string length: %v", err)
			}
			strBytes := make([]byte, strLen)
			if _, err := buf.Read(strBytes); err != nil {
				return Message{}, fmt.Errorf("failed to read string Value: %v", err)
			}
			payload.Value = string(strBytes)
		default:
			return Message{}, fmt.Errorf("unsupported Value type identifier: %v", valueType)
		}
	}

	message := Message{
		Header:  header,
		Payload: payload,
	}

	return message, nil
}
