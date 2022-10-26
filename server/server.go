package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	t "time"

	"chat/chat"

	"google.golang.org/grpc"
)

type Server struct {
	chat.UnimplementedChatServer
	name                             string
	port                             int
}

var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	flag.Parse()

	// Create a server struct
	server := &Server{
		name: "serverName",
		port: *port,
	}

	// Start the server
	go startServer(server)

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer(server *Server) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	list, err := net.Listen("tcp", ":9080")

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

func (s *Server) Chat(ctx context.Context, in *chat.RequestMsg) (*chat.ResponseMsg, error) {
	fmt.Printf("Received GetTime request\n")

	return &chat.ResponseMsg{
		: t.Now().String(),
	}, nil
}

func (s *Server) GetTime(ctx context.Context, in *time.GetTimeRequest) (*time.GetTimeReply, error) {
	fmt.Printf("Received GetTime request\n")

	return &time.GetTimeReply{
		Reply: t.Now().String(),
	}, nil
}