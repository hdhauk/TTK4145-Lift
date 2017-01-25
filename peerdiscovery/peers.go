package peerdiscovery

import (
	"fmt"
	"net"
	"sort"
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

// HeartBeatBeacon broadcast the supplied id every 15ms or whenever recieving
// a value on the transmitEnable channel.
func HeartBeatBeacon(port int, id string, transmitEnable <-chan bool) {

	conn := DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

// Start initiate listening for other peers while also start broadcasting
// to others
func Start(port int, onNewPeer func(IP string), onLostPeer func(IP string)) {
	var buf [1024]byte
	//var p PeerUpdate
	//lastSeen := make(map[string]time.Time)
	peers := make(map[string]peer)
	conn := DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		if id != "" {
			if _, idExists := peers[id]; !idExists {
				// Previusly unknown host
				peers[id] = peer{ip:}
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}
