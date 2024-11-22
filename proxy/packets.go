package proxy

import (
	"encoding/json"
	"fmt"

	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int           `json:"max"`
		Online int           `json:"online"`
		Sample []interface{} `json:"sample"`
	} `json:"players"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type PacketHandshake struct {
	ProtocolVersion     int32
	ServerAddress       string
	ServerPort          uint16
	EnumConnectionState int32
}

type PacketLoginStart struct {
	Name packet.String
	Uuid packet.UUID
}

type PacketStatusResponse struct {
	Response string
}

func WriteStatusResponse(conn *net.Conn, p PacketStatusResponse) error {
	return conn.WritePacket(packet.Marshal(
		0x00,
		packet.String(p.Response),
	))
}

func WriteHandshake(conn *net.Conn, p PacketHandshake) error {
	pkt := packet.Marshal(
		0x00,
		packet.VarInt(p.ProtocolVersion),
		packet.String(p.ServerAddress),
		packet.Short(p.ServerPort),
		packet.VarInt(p.EnumConnectionState),
	)
	return conn.WritePacket(pkt)
}

func WriteDisconnect(conn *net.Conn, reason string) error {
	return conn.WritePacket(packet.Marshal(
		0x00,
		packet.String(fmt.Sprintf("{\"text\":\"%s\"}", reason)),
	))
}

func WriteLoginStart(conn *net.Conn, p PacketLoginStart) error {
	pkt := packet.Marshal(
		0x00,
		p.Name,
		p.Uuid,
	)
	return conn.WritePacket(pkt)
}

func ReadStatusResponse(conn *net.Conn) (*StatusResponse, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet StatusResponse, got %d", p.ID)
	}

	var payload packet.String
	err = p.Scan(&payload)
	if err != nil {
		return nil, err
	}

	var resp StatusResponse
	err = json.Unmarshal([]byte(payload), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func ReadLoginStart(conn *net.Conn) (*PacketLoginStart, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet LoginStart, got %d", p.ID)
	}

	var name packet.String
	var uuid packet.UUID
	err = p.Scan(&name, &uuid)
	if err != nil {
		return nil, err
	}

	return &PacketLoginStart{
		Name: name,
		Uuid: uuid,
	}, nil
}

func ReadHandshake(conn *net.Conn) (*PacketHandshake, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet Handshake, got %d", p.ID)
	}

	var (
		protocolVersion packet.VarInt
		serverAddr      packet.String
		serverPort      packet.Short
		connectionState packet.VarInt
	)

	err = p.Scan(
		&protocolVersion,
		&serverAddr,
		&serverPort,
		&connectionState,
	)
	if err != nil {
		return nil, err
	}

	return &PacketHandshake{
		ProtocolVersion:     int32(protocolVersion),
		ServerAddress:       string(serverAddr),
		ServerPort:          uint16(serverPort),
		EnumConnectionState: int32(connectionState),
	}, nil
}
