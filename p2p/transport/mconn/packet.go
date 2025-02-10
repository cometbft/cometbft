package mconn

import (
	"encoding/binary"
	"io"
)

const (
	packetTypeMsg  = byte(0x01)
	packetTypePing = byte(0x02)
	packetTypePong = byte(0x03)
)

type Packet struct {
	ChID byte
	Type byte
	Data []byte
}

func (p *Packet) writeTo(w io.Writer) error {
	var err error
	err = binary.Write(w, binary.BigEndian, p.ChID)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, p.Type)
	if err != nil {
		return err
	}

	var length = uint32(len(p.Data))
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return err
	}

	_, err = w.Write(p.Data)
	return err
}

func readPacket(r io.Reader) (*Packet, error) {
	var chID, pType byte
	var length uint32

	err := binary.Read(r, binary.BigEndian, &chID)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.BigEndian, &pType)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	return &Packet{
		ChID: chID,
		Type: pType,
		Data: data,
	}, nil
}
