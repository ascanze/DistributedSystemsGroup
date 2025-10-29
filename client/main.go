package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	pb "DistributedSystemsGroup/grpc" // replace with your module path

	"google.golang.org/grpc"
)

// -----------------------------------------------------
// Client struct
// -----------------------------------------------------
type Client struct {
	id           string
	lamportClock int64
	mu           sync.Mutex
	service      pb.ChitChatClient
}

// -----------------------------------------------------
// Update Lamport clock only when receiving a broadcast
// -----------------------------------------------------
func (c *Client) updateClock(received int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if received > c.lamportClock {
		c.lamportClock = received
	}
	c.lamportClock++ // internal logical progression
}

// -----------------------------------------------------
// Main
// -----------------------------------------------------
func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	service := pb.NewChitChatClient(conn)
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)
	log.SetFlags(log.LstdFlags)
	log.SetPrefix(fmt.Sprintf("[CLIENT %s] ", id))

	client := &Client{id: id, lamportClock: 0, service: service}

	// --- Join the chat (no local Lamport increment) ---
	stream, err := service.Join(context.Background(), &pb.JoinRequest{Id: id})
	if err != nil {
		log.Fatalf("Could not join chat: %v", err)
	}

	// --- Goroutine: handle server broadcasts ---
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("[INFO] Server closed connection: %v", err)
				os.Exit(0)
			}

			// Update local Lamport clock
			client.updateClock(msg.GetLamportClock())

			// Display exactly what server broadcasted
			switch msg.GetType() {
			case pb.Broadcast_JOIN:
				log.Printf("[EVENT] Join | User=%s | Lamport=%d", msg.GetClientId(), msg.GetLamportClock())
			case pb.Broadcast_CHAT:
				log.Printf("[EVENT] Message | User=%s | Lamport=%d | Text=\"%s\"",
					msg.GetClientId(), msg.GetLamportClock(), msg.GetMessage())
			case pb.Broadcast_LEAVE:
				log.Printf("[EVENT] Leave | User=%s | Lamport=%d", msg.GetClientId(), msg.GetLamportClock())
			}
		}
	}()

	// --- Input loop for sending messages ---
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		if strings.ToLower(text) == "/exit" {
			_, err := service.Leave(context.Background(), &pb.LeaveRequest{ClientId: id})
			if err != nil {
				log.Printf("Error leaving chat: %v", err)
			}
			os.Exit(0)
		}

		_, err := service.Publish(context.Background(), &pb.PublishRequest{
			ClientId: id,
			Text:     text,
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
