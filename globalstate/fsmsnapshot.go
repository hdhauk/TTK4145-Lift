package globalstate

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	store State
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink
		if err := sink.Close(); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *fsmSnapshot) Release() {
	log.Println("[INFO] Release(): Finished with the snapshot")
}
