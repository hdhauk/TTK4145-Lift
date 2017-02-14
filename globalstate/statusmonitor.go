package globalstate

import (
	"time"

	"github.com/hashicorp/raft"
)

func (rw *raftwrapper) ConsensusMonitor() {
	lastStatus := raft.Candidate
	connected := false
	setConn := func(b bool) { connected = b }
	for {
		select {
		// Sleep slightly longer than the half the raft election timeout.
		case <-time.After(550 * time.Millisecond):
			newStatus := rw.raft.State()
			switch newStatus {
			case raft.Candidate:
				if lastStatus == raft.Candidate && connected {
					rw.config.OnLostConsensus()
					setConn(false)
				}
			case raft.Shutdown:
				rw.logger.Fatalf("[ERROR] Raft in shutdown!")
			default:
				if lastStatus == raft.Candidate {
					rw.config.OnAquiredConsensus()
				}
				setConn(true)
			}
			lastStatus = newStatus
		case <-rw.shutdown:
			return
		}
	}
}
