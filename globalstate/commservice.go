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
// The communication service is implmeneted with this interface instead of the
// actual store-type to limit access to private datamembers, and clearify which
// functions that is in use.
type StoreInterface interface {
	// Join joins the node, reachable at addr, to the cluster.
	Join(addr string) error

	// GetLeader returns the address of the current cluster leader
	GetLeader() string

	// GetStatus returns the current
	GetStatus() uint32

	UpdateLiftStatus(ls LiftStatus) error
	UpdateButtonStatus(bsu ButtonStatusUpdate) error
}

// service provides HTTP service for admitting new peers and command messages.
type service struct {
	addr       string
	leaderAddr string
	ln         net.Listener
	store      StoreInterface
}

func newCommService(addr string, store StoreInterface) *service {
	return &service{
		addr:  addr,
		store: store,
	}
}

// Start starts the communication service and start listening.
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

// Close closes the service.
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
	} else if strings.HasPrefix(p, "/update/lift") {
		s.HandleLiftUpdate(w, r)
		// Handle incomming updates
	} else if strings.HasPrefix(p, "/update/button") {
		s.HandleButtonUpdate(w, r)
	} else if strings.HasPrefix(p, "/cmd") {
		s.HandleCmd(w, r)
	} else if strings.HasPrefix(p, "/debug/dump-state") {
		s.HandleDebugDumpState(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		theFSM.logger.Printf("not found: someone tried to access %v", p)
	}
}

// Endpoint handlers
// =============================================================================
func (s *service) HandleJoin(w http.ResponseWriter, r *http.Request) {
	// Redirect if not currently leader
	if s.store.GetStatus() != 2 {
		// Infer commport from the raft-port. The communication port should always
		// be one above the raft port.
		leader := s.store.GetLeader()
		if leader == "" {
			theFSM.logger.Println("[WARN] No current leader. Cannot redirect")
			w.WriteHeader(http.StatusGone)
			return
		}
		parts := strings.Split(leader, ":")
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

func (s *service) HandleCmd(w http.ResponseWriter, r *http.Request) {
	// Check for empty reqest
	if r.Body == nil {
		http.Error(w, "No request body provided", http.StatusBadRequest)
		return
	}

	// Unmarshal json object (identical to btn-struct in LeaderMonitor
	// but redeclared here for your convenience ;)
	var btn = struct {
		Floor int
		Dir   string
	}{}

	err := json.NewDecoder(r.Body).Decode(&btn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	theFSM.config.OnIncomingCommand(btn.Floor, btn.Dir)
}

func (s *service) HandleLiftUpdate(w http.ResponseWriter, r *http.Request) {
	// Check for empty reqest
	if r.Body == nil {
		http.Error(w, "No request body provided", http.StatusBadRequest)
		return
	}

	// Unmarshal json object
	var status LiftStatus
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateLiftStatus(status); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	theFSM.logger.Printf("[INFO] Successfully accepted lift status update.")
	w.WriteHeader(http.StatusOK)
}

func (s *service) HandleButtonUpdate(w http.ResponseWriter, r *http.Request) {
	// Check for empty reqest
	if r.Body == nil {
		theFSM.logger.Printf("[WARN] Recieved empty button status update.\n")
		http.Error(w, "No request body provided", http.StatusBadRequest)
		return
	}

	// Unmarshal json object
	var status ButtonStatusUpdate
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		theFSM.logger.Println("[WARN] Unable to unmarshal incoming status update")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateButtonStatus(status); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	theFSM.logger.Printf("[INFO] Successfully accepted button status update.")
	w.WriteHeader(http.StatusOK)
}

func (s *service) HandleDebugDumpState(w http.ResponseWriter, r *http.Request) {
	state, _ := GetState()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(state)
}

func (s *service) HandleKick(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandlerKick is not yet implemented")
}