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

// commService provides HTTP commService for admitting new peers and command messages.
type commService struct {
	addr       string
	leaderAddr string
	ln         net.Listener
	store      *raftwrapper
	logger     *log.Logger
}

func newCommService(addr string, store *raftwrapper) *commService {
	return &commService{
		addr:   addr,
		store:  store,
		logger: store.logger,
	}
}

// Start starts the communication service and start listening.
func (s *commService) Start() error {
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
	http.ListenAndServe(s.addr, s)

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
func (s *commService) Close() {
	s.ln.Close()
	return
}

// ServeHTTP defines the behaviour when receiving a request
func (s *commService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	// Mux different endpoints
	if strings.HasPrefix(p, "/join") {
		// Join requests
		s.HandleJoin(w, r)
	} else if strings.HasPrefix(p, "/update/lift") {
		// Incomming lift status updates
		s.HandleLiftUpdate(w, r)
	} else if strings.HasPrefix(p, "/update/button") {
		// Incomming button status updates
		s.HandleButtonUpdate(w, r)
	} else if strings.HasPrefix(p, "/cmd") {
		// Incomming commands/assignments from leader
		s.HandleCmd(w, r)
	} else if strings.HasPrefix(p, "/debug/dump-state") {
		// For debugging purposes
		s.HandleDebugDumpState(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Printf("not found: someone tried to access %v", p)
	}
}

// Endpoint handlers
// =============================================================================
func (s *commService) HandleJoin(w http.ResponseWriter, r *http.Request) {
	// Redirect if not currently leader
	if s.store.GetStatus() != 2 {
		// Infer commport from the raft-port. The communication port should always
		// be one above the raft port.
		leader := s.store.GetLeader()
		if leader == "" {
			s.logger.Println("[WARN] No current leader. Cannot redirect")
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

func (s *commService) HandleCmd(w http.ResponseWriter, r *http.Request) {
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
	s.store.config.OnIncomingCommand(btn.Floor, btn.Dir)
}

func (s *commService) HandleLiftUpdate(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(http.StatusOK)
}

func (s *commService) HandleButtonUpdate(w http.ResponseWriter, r *http.Request) {
	// Check for empty reqest
	if r.Body == nil {
		s.logger.Printf("[WARN] Received empty button status update.\n")
		http.Error(w, "No request body provided", http.StatusBadRequest)
		return
	}

	// Unmarshal json object
	var status ButtonStatusUpdate
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		s.logger.Println("[WARN] Unable to unmarshal incoming status update")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateButtonStatus(status); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *commService) HandleDebugDumpState(w http.ResponseWriter, r *http.Request) {
	state := s.store.GetState()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(state)
}

func (s *commService) HandleKick(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandlerKick is not yet implemented")
}
