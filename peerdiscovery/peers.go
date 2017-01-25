package peerdiscovery

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// PeerUpdate contain a summary of changes in the list of known peers. All
// current peers are listed in Peers, while any new or lost peers are listed
// in New or Lost respectivly. Multiple lost peers usually indicate some sort of
// network failure.
type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

type peer struct {
	ip        string
	name      string
	firstSeen time.Time
	lastSeen  time.Time
}

const interval = 15 * time.Millisecond
const timeout = 50 * time.Millisecond

// broadcastHeartBeats broadcast the supplied id every 15ms or whenever recieving
// a value on the transmitEnable channel.
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
func Start(port int, ownID string, onNewPeer func(id, IP string)) {
	var buf [1024]byte
	peers := make(map[string]*peer)
	conn := dialBroadcastUDP(port)

	go broadcastHeartBeats(port, ownID)

	for {
		conn.SetReadDeadline(time.Now().Add(interval))

		// Although it is considered BAD go-code to throw away the error as we do
		// here the ReadFrom function will constantly yield non-nil error value
		// whenever nothing is read. Therefore we insted check to see if the string
		// n is empty further down.
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n]) // Either "" or on the form: "peerName@xxx.xxx.xxx.xxx"

		// TODO: Stop function from triggering on own heartbeats

		// Adding new connection
		if id != "" {
			if _, idExists := peers[id]; !idExists {
				// Previusly unknown host
				s := strings.Split(id, "@")
				peers[id] = &peer{
					ip:        s[1],
					name:      s[0],
					firstSeen: time.Now(),
					lastSeen:  time.Now(),
				}
				onNewPeer(s[0], s[1])
			}
			peers[id].lastSeen = time.Now()
		}

		// Removing dead connection
		// p.Lost = make([]string, 0)
		// for k, v := range lastSeen {
		// 	if time.Now().Sub(v) > timeout {
		// 		updated = true
		// 		p.Lost = append(p.Lost, k)
		// 		delete(lastSeen, k)
		// 	}
		// }

		// Sending update
		// if updated {
		// 	p.Peers = make([]string, 0, len(lastSeen))
		//
		// 	for k := range lastSeen {
		// 		p.Peers = append(p.Peers, k)
		// 	}
		//
		// 	sort.Strings(p.Peers)
		// 	sort.Strings(p.Lost)
		// 	peerUpdateCh <- p
		// }
	}
}
