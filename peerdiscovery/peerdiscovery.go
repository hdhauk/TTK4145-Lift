/*
Package peerdiscovery provides automatic detection of other peers in the same subnet.
It does this by utlizing broadcastmessages over UDP.
*/
package peerdiscovery

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Peer defines a peer
type Peer struct {
	IP        string
	Nick      string
	JoinPort  string
	RaftPort  string
	firstSeen time.Time
	lastSeen  time.Time
}

// Copy returns a copy of the peer
func (p *Peer) Copy() Peer {
	new := Peer{
		IP:       p.IP,
		Nick:     p.Nick,
		JoinPort: p.JoinPort,
		RaftPort: p.RaftPort,
	}
	return new
}

// Config defines all nessesary configuration parameters
type Config struct {
	Nick              string
	RaftPort          int
	BroadcastPort     int
	OnNewPeer         func(Peer)
	OnLostPeer        func(Peer)
	BroadcastInterval time.Duration
	Timeout           time.Duration
	Logger            *log.Logger
}

// broadcastHeartBeats broadcast the supplied id every 15ms
func broadcastHeartBeats(c Config) {
	conn := dialBroadcastUDP(c.BroadcastPort)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", c.BroadcastPort))

	// Resolve own IP-address
	ip, err := GetLocalIP()
	if err != nil {
		c.Logger.Printf("[ERROR] Unable to resolve own IP. Check internet connection")
	}

	// Create broadcast message: "nick"@"ip-address":"raft-port"
	id := fmt.Sprintf("%s@%s:%d", c.Nick, ip, c.RaftPort)

	for {
		select {
		case <-time.After(c.BroadcastInterval):
		}
		conn.WriteTo([]byte(id), addr)
	}
}

// Start initiate listening for other peers while also start broadcasting
// to others. The `id`-field in the callbacks have the form:
//	"nick"@"ip-address":"raft-port"
// where the latter part is the IPv4 adress of the peer.
func Start(c Config) {
	// Set up storage for peers we discover
	peers := make(map[string]*Peer)

	// Bind the socket
	conn := dialBroadcastUDP(c.BroadcastPort)

	// Start broadcasting
	go broadcastHeartBeats(c)

	// Start listening for others broadcasts
	var buf [1024]byte
	for {
		conn.SetReadDeadline(time.Now().Add(c.BroadcastInterval))

		// Although it is considered BAD go-code to throw away the error as we do
		// here the ReadFrom function will constantly yield non-nil error value
		// whenever nothing is read. Therefore we instead check to see if the string
		// n is empty further down.
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n]) // Either "" or on the form: nick"@"ip-address":"raft-port"

		// Stop function from triggering on own heartbeats
		if strings.Contains(id, c.Nick) {
			continue
		}

		// Adding new connection
		if id != "" {
			if _, idExists := peers[id]; !idExists {
				// Previusly unknown host
				parts := strings.Split(id, "@")
				ipv4parts := strings.Split(parts[1], ":")
				peers[id] = &Peer{
					IP:        ipv4parts[0],
					Nick:      parts[0],
					RaftPort:  ipv4parts[1],
					firstSeen: time.Now(),
					lastSeen:  time.Now(),
				}
				c.OnNewPeer(peers[id].Copy())
			}
			peers[id].lastSeen = time.Now()
		}
	}
}
