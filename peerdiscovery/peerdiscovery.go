/*
Package peerdiscovery provides automatic detection of other peers in the same subnet.
It does this by utlizing broadcastmessages over UDP.
*/
package peerdiscovery

import (
	"fmt"
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

const interval = 15 * time.Millisecond
const timeout = 50 * time.Millisecond

// broadcastHeartBeats broadcast the supplied id every 15ms
func broadcastHeartBeats(port int, id string) {
	conn := dialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	for {
		select {
		case <-time.After(interval):
		}
		conn.WriteTo([]byte(id), addr)
	}
}

// Start initiate listening for other peers while also start broadcasting
// to others. The `id`-field in the callbacks have the form:
//	peerName@xxx.xxx.xxx.xxx
// where the latter part is the IPv4 adress of the peer.
func Start(port int, ownID, joinPort, raftPort string, onNewPeer func(peer Peer)) {
	var buf [1024]byte
	peers := make(map[string]*Peer)
	conn := dialBroadcastUDP(port)

	go broadcastHeartBeats(port, fmt.Sprintf("%s@%s@%s", ownID, joinPort, raftPort))

	for {
		conn.SetReadDeadline(time.Now().Add(interval))

		// Although it is considered BAD go-code to throw away the error as we do
		// here the ReadFrom function will constantly yield non-nil error value
		// whenever nothing is read. Therefore we instead check to see if the string
		// n is empty further down.
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n]) // Either "" or on the form: "peerName@xxx.xxx.xxx.xxx"

		// Stop function from triggering on own heartbeats
		if strings.Contains(id, ownID) {
			continue
		}

		// Adding new connection
		if id != "" {
			if _, idExists := peers[id]; !idExists {
				// Previusly unknown host
				s := strings.Split(id, "@")
				peers[id] = &Peer{
					IP:        s[1],
					Nick:      s[0],
					JoinPort:  s[2],
					RaftPort:  s[3],
					firstSeen: time.Now(),
					lastSeen:  time.Now(),
				}
				onNewPeer(peers[id].Copy())
			}
			peers[id].lastSeen = time.Now()
		}
	}
}
