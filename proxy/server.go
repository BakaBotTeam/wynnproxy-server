package proxy

import (
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"wynnproxyserver/icon"

	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

var (
	Users                 = make(map[string]User)
	Names                 = make(map[string]string)
	HANDSHAKE_STATUS_ID   = int32(1)
	HANDSHAKE_LOGIN_ID    = int32(2)
	HANDSHAKE_TRANSFER_ID = int32(3)
)

type User struct {
	ServerAddress string
}

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

	if handshake.EnumConnectionState == HANDSHAKE_STATUS_ID {
		err = server.handlePing(conn, *handshake)
		return err
	} else if handshake.EnumConnectionState == HANDSHAKE_LOGIN_ID {
		err = server.forwardConnection(conn, *handshake)
		return err
	} else if handshake.EnumConnectionState == HANDSHAKE_TRANSFER_ID {
		err = server.forwardHandshakeConnection(conn, *handshake)
		return err
	}

	return nil
}

// forward connection
func (server *MinecraftProxyServer) forwardHandshakeConnection(clientConn *mcnet.Conn, handshake PacketHandshake) error {
	value, nameExists := Names[strings.Split(clientConn.Socket.RemoteAddr().String(), ":")[0]]
	if nameExists {
		user, exists := Users[value]
		if exists {
			remoteConn, err := mcnet.DialMC(user.ServerAddress)
			if err != nil {
				return err
			}
			defer func() {
				remoteConn.Close()
			}()

			splitAddress := strings.SplitN(user.ServerAddress, ":", 2)
			port, _ := strconv.Atoi(splitAddress[1])
			handshake.ServerAddress = splitAddress[0]
			handshake.ServerPort = uint16(port)
			handshake.EnumConnectionState = HANDSHAKE_LOGIN_ID

			WriteHandshake(remoteConn, handshake)

			loginStart, err := ReadLoginStart(clientConn)
			if err != nil {
				return err
			}

			WriteLoginStart(remoteConn, *loginStart)

			log.Println("Forwarding packets for", string(loginStart.Name))

			var waitGroup sync.WaitGroup
			waitGroup.Add(2)

			go func() {
				io.Copy(remoteConn, clientConn)
				waitGroup.Done()
			}()

			go func() {
				io.Copy(clientConn, remoteConn)
				waitGroup.Done()
			}()

			waitGroup.Wait()
		}
	}
	return nil
}

// forward connection
func (server *MinecraftProxyServer) forwardConnection(clientConn *mcnet.Conn, handshake PacketHandshake) error {
	remoteConn, err := mcnet.DialMC(server.Remote)
	if err != nil {
		return err
	}
	defer func() {
		remoteConn.Close()
	}()

	splitAddress := strings.SplitN(server.Remote, ":", 2)
	port, _ := strconv.Atoi(splitAddress[1])
	handshake.ServerAddress = splitAddress[0]
	handshake.ServerPort = uint16(port)

	WriteHandshake(remoteConn, handshake)

	loginStart, err := ReadLoginStart(clientConn)
	if err != nil {
		return err
	}

	Names[strings.Split(clientConn.Socket.RemoteAddr().String(), ":")[0]] = string(loginStart.Name)

	WriteLoginStart(remoteConn, *loginStart)

	log.Println("Forwarding packets for", string(loginStart.Name), "| Waiting to transfer to proxy server...")

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		io.Copy(remoteConn, clientConn)
		waitGroup.Done()
	}()

	go func() {
		io.Copy(clientConn, remoteConn)
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
				Response: packet.String(bytes),
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
