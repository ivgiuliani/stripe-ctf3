package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goraft/raft"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"stripe-ctf.com/sqlcluster/command"
	"stripe-ctf.com/sqlcluster/log"
	"stripe-ctf.com/sqlcluster/sql"
	"stripe-ctf.com/sqlcluster/transport"
	"time"
)

type Server struct {
	name       string
	path       string
	listen     string
	router     *mux.Router
	raftServer raft.Server
	httpServer *http.Server
	sql        *sql.SQL
	client     *transport.Client
}

// Creates a new server.
func New(path, listen string) (*Server, error) {
	s := &Server{
		name:   listen,
		path:   path,
		listen: listen,
		sql:    sql.NewSQL(),
		router: mux.NewRouter(),
		client: transport.NewClient(),
	}

	return s, nil
}

// Returns the connection string.
func (s *Server) connectionString() string {
	url, err := transport.Encode(s.listen)
	if err != nil {
		log.Fatal(err)
	}
	return url
}

// Starts the server.
func (s *Server) ListenAndServe(leader string) error {
	var err error

	// Initialize and start HTTP server.
	log.Println("Initializing HTTP server")
	s.httpServer = &http.Server{
		Handler: s.router,
	}

	s.router.HandleFunc("/sql", s.sqlHandler).Methods("POST")
	s.router.HandleFunc("/join", s.joinHandler).Methods("POST")

	// Start Unix transport
	log.Println(s.name, "listening at", s.listen)
	l, err := transport.Listen(s.listen)
	if err != nil {
		log.Fatal(err)
	}

	// initialize raft
	transporter := raft.NewHTTPTransporter("/raft")

	// swap the dialer with the unix dialer that also allows unix-sockets to
	// be passed around
	transporter.Transport.Dial = transport.UnixDialer

	s.raftServer, err = raft.NewServer(s.name, s.path, transporter, nil, s.sql, "")
	if err != nil {
		log.Fatal(err)
	}
	transporter.Install(s.raftServer, s)

	// this seems to yield good results, but it's definitely not the most
	// empirical of the measurements
	s.raftServer.SetElectionTimeout(800 * time.Millisecond)
	s.raftServer.SetHeartbeatTimeout(150 * time.Millisecond)
	s.raftServer.Start()

	if leader != "" {
		// Join the leader if specified.
		log.Println("Attempting to join the leader:", leader)
		time.Sleep(500 * time.Millisecond)

		if !s.raftServer.IsLogEmpty() {
			log.Fatal("Cannot join with an existing log")
			return nil
		}

		// retry the join until we actually join the cluster
		// (it may take a while, until octopus sets the sockets up)
		for {
			if err := s.Join(leader); err == nil {
				break
			}
			log.Fatal(err)
			time.Sleep(10 * time.Millisecond)
		}
		log.Println("joined.")
	} else if s.raftServer.IsLogEmpty() {
		// Initialize the server by joining itself.
		log.Println("Initializing new cluster")

		_, err := s.raftServer.Do(&raft.DefaultJoinCommand{
			Name:             s.raftServer.Name(),
			ConnectionString: s.connectionString(),
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Recovered from log")
	}

	log.Println(s.name, "IS READY TO ACCEPT REQUESTS")

	return s.httpServer.Serve(l)
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

// Joins to the leader of an existing cluster.
func (s *Server) Join(leader string) error {
	cs, err := transport.Encode(leader)
	if err != nil {
		log.Fatal(err)
	}

	command := &raft.DefaultJoinCommand{
		Name:             s.raftServer.Name(),
		ConnectionString: s.connectionString(),
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(command)
	log.Println(fmt.Sprintf("%s is joining cluster %s from %s (connection string = %s)", s.name, leader, s.listen, cs))
	_, err = s.client.SafePost(cs, "/join", &b)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
	join_command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(req.Body).Decode(&join_command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var nodename string = join_command.NodeName()
	log.Println(nodename, "from", join_command.ConnectionString, "asked to join the cluster")

	if _, err := s.raftServer.Do(join_command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(nodename, "has joined the cluster.")
}

// This is the only user-facing function, and accordingly the body is
// a raw string rather than JSON.
func (s *Server) sqlHandler(w http.ResponseWriter, req *http.Request) {
	// read the query
	query, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.raftServer.State() == raft.Leader {
		// just execute the query if we're the leader
		resp, err := s.execSQL(string(query))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write(resp)
		return
	}

	// TODO: there must be a better way to do all this

	// get a list of available mates
	var _mates []string
	var mates []string
	_mates = make([]string, 0, len(s.raftServer.Peers())+1)
	mates = make([]string, 0, len(s.raftServer.Peers())+1)

	for _, peer := range s.raftServer.Peers() {
		_mates = append(_mates, peer.Name)
		mates = append(mates, peer.Name)
	}

	// shuffle the hosts
	perm := rand.Perm(len(_mates))
	for i, v := range perm {
		mates[v] = _mates[i]
	}

	// if we know the leader, always try it first and only fallback
	// to other hosts in the cluster if we don't (but maybe they do)
	if s.raftServer.Leader() != "" {
		mates = append(mates, s.raftServer.Leader())
	}

	// reverse the peers so that the leader is always first
	for i, j := 0, len(mates)-1; i < j; i, j = i+1, j-1 {
		mates[i], mates[j] = mates[j], mates[i]
	}

	for _, who := range mates {
		// proxy to whoever is available
		cs, err := transport.Encode(who)
		if err != nil {
			continue
		}

		// the number of retries is totally arbitrary
		if output, ok := s.proxy(cs, query, len(mates)); ok {
			w.Write(output)
			return
		}
	}

	http.Error(w, "error", http.StatusBadRequest)
}

func (s *Server) proxy(whoCs string, query []byte, retries int) ([]byte, bool) {
	var output []byte

	queryb := bytes.NewBufferString(string(query))
	resp, err := s.client.SafePost(whoCs, "/sql", queryb)
	if err != nil {
		//log.Println("cannot get to the leader, retrying")
		return nil, false
	}

	output, err = ioutil.ReadAll(resp)
	var ok bool
	if output == nil {
		// just wait for the query to be replicated, it will eventually
		// turn up here, but if it takes too long just abort the try
		for i := 0; i < retries; i++ {
			output, ok = s.getQueryOutput(query)
			if ok {
				break
			}

			// just wait a bit before trying again
			time.Sleep(15 * time.Millisecond)
		}
	}

	return output, true
}

func (s *Server) getQueryOutput(query []byte) ([]byte, bool) {
	key, _, _ := s.sql.GetData(string(query))
	output, err := s.sql.ReadOutput(key)
	if err != nil {
		return nil, false
	}
	return output, true
}

func (s *Server) execSQL(query string) ([]byte, error) {
	var create bool = false
	var name, word string
	var count int

	if strings.HasPrefix(query, "CREATE") {
		create = true
	} else {
		word, name, count = s.sql.GetData(query)
	}

	resp, err := s.raftServer.Do(command.NewQueryCommand(create, name, word, count))
	if err != nil {
		return nil, err
	}

	return resp.([]byte), nil
}
