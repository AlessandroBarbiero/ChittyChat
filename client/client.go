package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"

	"github.com/AlessandroBarbiero/Dist-Sys-Assignment4/time"

	"google.golang.org/grpc"
)

func main() {
	// Creat a virtual RPC Client Connection on port  9080 WithInsecure (because  of http)
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}

	// Defer means: When this function returns, call this method (meaning, one main is done, close connection)
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	c := time.NewGetCurrentTimeClient(conn)

	reader := bufio.NewReader(os.Stdin)

	//Send a request when something is written on the console
	for {
		_, _, err := reader.ReadLine()
		if err != nil {
			log.Fatalf("Cannot read from input")
		}
		SendGetTimeRequest(c)
		//t.Sleep(5 * t.Second)
	}
}

func SendGetTimeRequest(c time.GetCurrentTimeClient) {
	// Between the curly brackets are nothing, because the .proto file expects no input.
	message := time.GetTimeRequest{}

	response, err := c.GetTime(context.Background(), &message)
	if err != nil {
		log.Fatalf("Error when calling GetTime: %s", err)
	}

	fmt.Printf("Current time right now: %s\n", response.Reply)
}
