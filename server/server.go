package main

import (
	"chat/chat"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

type Server struct {
	chat.UnimplementedChatServer
	name string
	port int
	// Store a progressive number for the Ids of the clients in order to give a unique id to each client
	idCounter int64
	clients   map[int64]chat.Chat_ChatServer
	mutex     sync.RWMutex
}

// Add this part if we want to use parametric port on call of the method
// var port = flag.Int("port", 0, "server port number")

func main() {
	// Set the log file output
	f, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error trying to open/create the log file")
	}

	defer f.Close()

	log.SetOutput(f)

	// Get the port from the command line when the server is run
	// flag.Parse()

	// Hardcoded port
	portValue := 8080
	port := &portValue

	// Create a server struct
	server := &Server{
		name:      "serverName",
		port:      *port,
		idCounter: 0,
		clients:   make(map[int64]chat.Chat_ChatServer),
		mutex:     sync.RWMutex{},
	}

	// Start the server
	startServer(server)
}

func startServer(server *Server) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	list, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}

	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	chat.RegisterChatServer(grpcServer, server)
	serveError := grpcServer.Serve(list)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (s *Server) Chat(stream chat.Chat_ChatServer) error {
	log.Printf("Received Join Request\n")

	id := s.addClient(stream)
	// At the end of the method remove the client
	defer s.removeClient(id)

	// Send a message from the server stating that a client joined the chat
	msg := chat.ResponseMsg{
		Id:          0,
		Message:     "Client " + strconv.Itoa(int(id)) + " joined the chat",
		VectorClock: make(map[int64]int64),
	}

	msg.VectorClock[id] = 0

	// Send the joining message also to the joining client
	s.broadcastMessage(&msg)

	for {
		// Wait for msg from client
		req, err := stream.Recv()
		if err == io.EOF {
			// Client disconnected
			break
		}

		if err != nil {
			log.Printf("Server couldn't receive message from client %d.\nError message: %s\n", id, err)
			// Client disconnected
			break
		}

		// Send message to every other client connected
		msg.Id = req.Id
		msg.Message = req.Message
		msg.VectorClock = req.VectorClock

		s.broadcastMessage(&msg)
	}

	return nil
}

// Add a new client with an incremented idCounter, send a message to him to set his id
// and return the new id assigned to him
func (s *Server) addClient(stream chat.Chat_ChatServer) int64 {

	var id int64

	s.mutex.Lock()
	s.idCounter++
	id = s.idCounter
	s.clients[id] = stream
	s.mutex.Unlock()

	// send a direct message to the client to assign his id
	msg := chat.ResponseMsg{
		Id:          id,
		Message:     "This is your id",
		VectorClock: nil,
	}
	stream.Send(&msg)

	log.Printf("Client %d has joined the chat", s.idCounter)
	return id
}

// Remove the client from the list of saved clients connected to the server and notify the others.
// If the client is not present this is a no-op
func (s *Server) removeClient(id int64) {

	// Delete the client and send message only if the client is present
	_, ok := s.clients[id]
	if ok {
		s.mutex.Lock()
		delete(s.clients, id)
		s.mutex.Unlock()

		removeMsg := chat.ResponseMsg{
			Id:          id,
			Message:     "Client " + strconv.Itoa(int(id)) + " left the chat",
			VectorClock: make(map[int64]int64),
		}
		s.broadcastMessage(&removeMsg)
		log.Printf("Client %d left the chat.\n", id)
	}
}

// Broadcast message to every client connected to the server except the sender,
// if the id of the sender is 0 the message is coming from the server.
// If the message cannot be sent to a client disconnect the client
func (s *Server) broadcastMessage(msg *chat.ResponseMsg) {
	for id, stream := range s.clients {

		if id != msg.Id {
			err := stream.Send(msg)
			if err != nil {
				log.Printf("Can't send message to client %d\nRemove it from active clients", id)
				s.removeClient(id)
			}
		}
	}
	if msg.Id == 0 {
		log.Printf("Message from server was broadcasted.\n")
	} else {
		log.Printf("Message from client %d was broadcasted.\n", msg.Id)
	}
}
