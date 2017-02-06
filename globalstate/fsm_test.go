package globalstate

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_FSMStart(t *testing.T) {
	f1 := newFSM("")
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	f1.RaftPort = "0"
	f1.RaftDir = tmpDir

	if f1 == nil {
		t.Fatalf("failed to create store")
	}

	if err := f1.Start(false); err != nil {
		t.Fatalf("failed to start store: %s", err)
	}
}
