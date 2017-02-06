package globalstate

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// StoreInterface is the interface Raft-backed key-value stores must implement.
type StoreInterface interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// GetLeader returns the address of the current cluster leader
	GetLeader() string

	// GetStatus TODO
	GetStatus() uint32
}

// service provides HTTP service for admitting new peers and command messages.
type service struct {
	addr       string
	leaderAddr string
	ln         net.Listener

	store StoreInterface
}

func newCommService(addr string, store StoreInterface) *service {
	return &service{
		addr:  addr,
		store: store,
	}
}

func (s *service) Start() error {
	server := http.Server{
		Handler: s,
	}

	// Create listener
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = l

	// Handle request to the root
	http.Handle("/", s)

	// Start accepting incomming connections on the listener
	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()
	return nil
}

// close closes the service.
func (s *service) Close() {
	s.ln.Close()
	return
}

// ServeHTTP defines the behaviour when receiving a request
func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	// Mux different endpoints
	if strings.HasPrefix(p, "/join") {
		// Handle join requests
		s.HandleJoin(w, r)
	} else if strings.HasPrefix(p, "/key") {
		// Handle incomming key update
	} else if strings.HasPrefix(p, "/cmd") {
		// Handle incomming command
	} else {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("not found: someone tried to access %v", p)
	}
}

func (s *service) HandleJoin(w http.ResponseWriter, r *http.Request) {
	// Redirect if not currently leader
	if s.store.GetStatus() != 2 {
		// Infer commport from the raft-port. The communication port should always
		// be one above the raft port.
		parts := strings.Split(s.store.GetLeader(), ":")
		raftPortInt, _ := strconv.Atoi(parts[1])
		portStr := strconv.Itoa(raftPortInt + 1)

		// Return leader address to requestor
		w.Header().Add("X-Raft-Leader", parts[0]+":"+portStr)
		return
	}

	// Decode incoming json object. Discard if unable to decode
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("Unable to decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Simple and naive test to prevent injection of more than one peer
	if len(m) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Pull port off request
	peerAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	parts := strings.Split(r.RemoteAddr, ":")
	peerAddr = parts[0] + ":" + peerAddr

	if err := s.store.Join(peerAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Return empty X-Raft-Leader tag to indicate success
	w.Header().Add("X-Raft-Leader", "")
	w.WriteHeader(http.StatusOK)
}

func (s *service) HandleKick(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandlerKick is not yet implemented")
}

func (s *service) HandleCmd(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleCmd not implemented yet.")
}

func (s *service) HandleKey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleKey not implemented yet.")

}

// Addr returns the address on which the Service is listening
func (s *service) Addr() net.Addr {
	return s.ln.Addr()
}
