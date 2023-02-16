package webrtc

import (
	"log"
	"sync"

	"github.com/gofiber/websocket"
)

func RoomConn(c *websocket.Conn, p *Peers) {
	var config webrtc.configuration

	peerConnection, err := webrtc.NewPeerConnection(config)

	if err != nil {
		log.Print(err)
		return
	}

	newPeer := PeerConnectionState{
		PeerConnection: peerConnection,
		websocket:      &ThreadSafeWriter{},
		Conn:           c,
		Mutex:          sync.Mutex{},
	}

}
