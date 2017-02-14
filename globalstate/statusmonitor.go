package globalstate

import (
	"time"

	"github.com/hashicorp/raft"
)

func (f *raftwrapper) ConsensusMonitor() {
	lastStatus := raft.Candidate
	connected := false
	setConn := func(b bool) { connected = b }
	for {
		newStatus := f.raft.State()
		switch newStatus {
		case raft.Candidate:
			if lastStatus == raft.Candidate && connected {
				f.config.OnLostConsensus()
				setConn(false)
			}
		case raft.Shutdown:
			f.logger.Fatalf("[ERROR] Raft in shutdown!")
		default:
			if lastStatus == raft.Candidate {
				f.config.OnAquiredConsensus()
			}
			setConn(true)
		}
		lastStatus = newStatus
		// Sleep slightly longer than the half the raft election timeout.
		time.Sleep(550 * time.Millisecond)
	}
}
