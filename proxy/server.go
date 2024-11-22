package proxy

import (
	"encoding/json"
	"io"
	"log"
	"strings"
	"sync"
	"wynnproxyserver/icon"

	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

type MinecraftProxyServer struct {
	Listen string
	Remote string
	MOTD   string

	running  bool
	listener *mcnet.Listener
}

func (server *MinecraftProxyServer) StartServer() error {
	var err error
	server.listener, err = mcnet.ListenMC(server.Listen)
	if err != nil {
		return err
	}
	server.running = true

	for server.running {
		conn, err := server.listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			server.handleConnection(&conn)
		}()
	}

	return nil
}

func (server *MinecraftProxyServer) CloseServer() {
	if server.running {
		server.listener.Close()
	}
}

func (server *MinecraftProxyServer) handleConnection(conn *mcnet.Conn) error {
	defer conn.Close()
	handshake, err := ReadHandshake(conn)

	if err != nil {
		return err
	}

	if handshake.EnumConnectionState == 1 { //protocol id 1: Status
		err = server.handlePing(conn, *handshake)
		return err
	} else if handshake.EnumConnectionState == 2 { //protocol id 2: Login
		err = server.forwardConnection(conn, *handshake)
		return err
	}

	return nil
}

// forward connection
func (s *MinecraftProxyServer) forwardConnection(conn *mcnet.Conn, handshake PacketHandshake) error {
	remoteConn, err := mcnet.DialMC(s.Remote)
	if err != nil {
		return err
	}
	defer remoteConn.Close()

	handshake.ServerAddress = strings.SplitN(s.Remote, ":", 2)[0]
	handshake.ServerPort = uint16(25565)

	WriteHandshake(remoteConn, handshake)

	loginStart, err := ReadLoginStart(conn)
	if err != nil {
		return err
	}

	WriteLoginStart(remoteConn, *loginStart)

	log.Println("Forwarding packets for", string(loginStart.Name))

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		io.Copy(remoteConn, conn)
		waitGroup.Done()
	}()

	go func() {
		io.Copy(conn, remoteConn)
		waitGroup.Done()
	}()

	waitGroup.Wait()

	return nil
}

// handle ping request
func (server *MinecraftProxyServer) handlePing(conn *mcnet.Conn, handshake PacketHandshake) error {
	for {
		var pkt packet.Packet
		err := conn.ReadPacket(&pkt)
		if err != nil {
			return err
		}

		switch pkt.ID {
		case 0x00: // status request
			resp := StatusResponse{}

			resp.Version.Name = "WynnCraft"
			resp.Version.Protocol = int(handshake.ProtocolVersion)
			resp.Players.Max = 1919810
			resp.Players.Online = 114514
			resp.Players.Sample = []interface{}{
				"WynnCraft",
			}
			resp.Favicon = "data:image/png;base64," + icon.Favicon
			resp.Description = server.MOTD

			bytes, err := json.Marshal(resp)
			if err != nil {
				return nil
			}

			err = WriteStatusResponse(conn, PacketStatusResponse{
				Response: string(bytes),
			})

			if err != nil {
				return err
			}

		case 0x01: // ping
			var payload packet.Long
			err := pkt.Scan(&payload)
			if err != nil {
				return err
			}

			err = conn.WritePacket(packet.Marshal(
				0x01,
				payload),
			)

			if err != nil {
				return err
			}
		}
	}

}
