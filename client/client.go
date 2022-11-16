package main

import (
	"bufio"
	"chat/chat"
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var MaxLength = 128
var PORT = 8080
var IP = "localhost"

type Client struct {
	id          int64
	vectorClock map[int64]int64
	mutex       sync.RWMutex
}

// Add this part if we want to use parametric port on command line
// var serverPort = flag.Int("serverPort", 8080, "server port number")  //name, port, usage OK

func main() {
	// Get the port from the command line when the client is run
	// flag.Parse()

	// Hardcoded port
	portValue := PORT
	serverPort := &portValue

	client := &Client{
		id:          1,
		vectorClock: make(map[int64]int64),
		mutex:       sync.RWMutex{},
	}

	client.startClient(serverPort)
}

// Get connection to server
// Start receiving thread
// Wait for user input and send the messages to server
func (c *Client) startClient(serverPort *int) {
	serverConnection := getServerConnection(serverPort)

	ctx := context.Background()
	// Request the server to join the chat
	stream, err := serverConnection.Chat(ctx)
	if err != nil {
		log.Fatalln("Client couldn't connect")
	}

	// Receive the first message from the server to know my id
	msg, err := stream.Recv()
	if err != nil {
		log.Fatalln("Couldn't get id from server")
	}

	// Set the log file with the right client id
	f, err := os.OpenFile("client"+strconv.Itoa(int(msg.Id))+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error trying to open/create the log file")
	}

	defer f.Close()

	log.SetOutput(f)

	c.id = msg.Id
	c.vectorClock[c.id] = 0

	// start receiving thread for all messages
	go c.receiveMessages(stream)

	// Wait for user input to send messages, if the message is longer than MaxLength it is not sent
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if len(input) > MaxLength {
			log.Println("Invalid message, please keep length of message under 128 characters")
		} else {
			c.sendMessage(stream, input)
			log.Printf("Message sent:\"%s\"\n", input)
		}
	}
}

// Receive the message from the stream,
// Print it with the vectorClock sent,
// Update my vectorClock based on the received one
// Increase my own clock and print it
func (c *Client) receiveMessages(stream chat.Chat_ChatClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("Disrupted connection")
		}

		if msg.Id == 0 {
			log.Printf("Received message from server: \"%s\" sent at %v", msg.Message, msg.VectorClock)
		} else {
			log.Printf("Received message from %d: \"%s\" sent at %v", msg.Id, msg.Message, msg.VectorClock)
		}

		// update the clock
		c.mergeVectorClocks(msg.VectorClock)
		c.updateMyTime()
		log.Printf("Current node time is %v", c.vectorClock)
	}
}

// +1 in vector clock for this node
func (c *Client) updateMyTime() map[int64]int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.vectorClock[c.id]++

	return c.vectorClock
}

// merging vector clocks, like it is defined in algorithm
func (c *Client) mergeVectorClocks(clock map[int64]int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for client, clientTime := range clock {
		_, ok := c.vectorClock[client]
		if ok {
			c.vectorClock[client] = Max(c.vectorClock[client], clientTime)
		} else {
			c.vectorClock[client] = clientTime
		}
	}
}

// Update the vector clock and send a message with the given content and the updated clock
func (c *Client) sendMessage(stream chat.Chat_ChatClient, content string) {

	clock := c.updateMyTime()

	msg := chat.RequestMsg{
		Id:          c.id,
		Message:     content,
		VectorClock: clock,
	}

	err := stream.SendMsg(&msg)

	if err != nil {
		log.Printf("Couldn't send message")
	}
}

// dial server
func getServerConnection(serverPort *int) chat.ChatClient {

	// Run server and client on local host if you don't write ip address
	conn, err := grpc.Dial(IP+":"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("could not dial")
	}
	log.Printf("Dialed\n")

	return chat.NewChatClient(conn)
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
