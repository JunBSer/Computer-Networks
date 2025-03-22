package models

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

type Message struct {
	SenderID   int32           `json:"sender_id"`
	ReceiverId int32           `json:"receiver_id"`
	Body       json.RawMessage `json:"body"`
}

func ParseMsg(data []byte) (*Message, error) {
	msg := &Message{}

	if err := json.Unmarshal(data, msg); err != nil {
		return &Message{}, err
	}

	return msg, nil
}

func GetMessage(conn net.Conn) ([]byte, error) {
	var msgLen int32
	err := binary.Read(conn, binary.LittleEndian, &msgLen)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, msgLen)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func SendMessage(conn net.Conn, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	msgLen := uint32(len(data))
	if err := binary.Write(conn, binary.LittleEndian, msgLen); err != nil {
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}
