package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	pb "DistributedSystemsGroup/grpc" // <-- replace with your actual module path if different

	"google.golang.org/grpc"
)

// -----------------------------------------------------
// ChitChatServer: holds state for active clients + Lamport
// -----------------------------------------------------
type ChitChatServer struct {
	pb.UnimplementedChitChatServer

	mu      sync.Mutex
	clients map[string]pb.ChitChat_JoinServer // id -> server stream to that client
	lamport int64                             // Lamport logical clock
}

// nextLamport increments the server Lamport clock safely and returns it.
func (s *ChitChatServer) nextLamport() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lamport++
	return s.lamport
}

// -----------------------------------------------------
// Join: client opens a server-stream to receive broadcasts
// -----------------------------------------------------
func (s *ChitChatServer) Join(req *pb.JoinRequest, stream pb.ChitChat_JoinServer) error {
	id := req.GetId()

	// Store this client's stream
	s.mu.Lock()
	if s.clients == nil {
		s.clients = make(map[string]pb.ChitChat_JoinServer)
	}
	s.clients[id] = stream
	s.mu.Unlock()

	// Connection log
	log.Printf("[EVENT] Connection | User=%s", id)

	// Broadcast JOIN (with Lamport)
	L := s.nextLamport()
	log.Printf("[EVENT] Join | User=%s | Lamport=%d", id, L)
	s.broadcast(&pb.Broadcast{
		Type:         pb.Broadcast_JOIN,
		ClientId:     id,
		Message:      fmt.Sprintf("%s joined the chat", id),
		LamportClock: L,
	})

	// Keep the stream open until client disconnects or server shuts down.
	// We wait on the stream context; when client goes away, ctx.Done() is closed.
	<-stream.Context().Done()

	// If the client disappears without calling Leave(), remove its stream entry.
	s.mu.Lock()
	if _, ok := s.clients[id]; ok {
		delete(s.clients, id)
		// (Optional) Emit a disconnection event without Lamport bump,
		// since "Leave" is the canonical leave event. This line documents the drop.
		log.Printf("[EVENT] Disconnection | User=%s (stream closed)", id)
	}
	s.mu.Unlock()

	return stream.Context().Err()
}

// -----------------------------------------------------
// Publish: client sends a chat message (unary RPC)
// -----------------------------------------------------
func (s *ChitChatServer) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	id := req.GetClientId()
	text := req.GetText()

	// Validate UTF-8 and length ≤ 128 runes
	if !utf8.ValidString(text) {
		return &pb.PublishResponse{Ack: false, Error: "message is not valid UTF-8"}, nil
	}
	if len([]rune(text)) > 128 {
		return &pb.PublishResponse{Ack: false, Error: "message too long (max 128 chars)"}, nil
	}

	L := s.nextLamport()
	log.Printf("[EVENT] Message | User=%s | Lamport=%d | Text=\"%s\"", id, L, text)

	s.broadcast(&pb.Broadcast{
		Type:         pb.Broadcast_CHAT,
		ClientId:     id,
		Message:      text,
		LamportClock: L,
	})

	return &pb.PublishResponse{Ack: true}, nil
}

// -----------------------------------------------------
// Leave: client leaves gracefully (unary RPC)
// -----------------------------------------------------
func (s *ChitChatServer) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	id := req.GetClientId()

	// Remove from clients map
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()

	// Emit LEAVE with Lamport
	L := s.nextLamport()
	log.Printf("[EVENT] Leave | User=%s | Lamport=%d", id, L)

	s.broadcast(&pb.Broadcast{
		Type:         pb.Broadcast_LEAVE,
		ClientId:     id,
		Message:      fmt.Sprintf("%s left the chat", id),
		LamportClock: L,
	})

	// Disconnection log (explicit)
	log.Printf("[EVENT] Disconnection | User=%s", id)

	return &pb.LeaveResponse{Ack: true}, nil
}

// -----------------------------------------------------
// broadcast: fan-out a broadcast to all active client streams
// -----------------------------------------------------
func (s *ChitChatServer) broadcast(b *pb.Broadcast) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, stream := range s.clients {
		if err := stream.Send(b); err != nil {
			log.Printf("[ERROR] Send failure to %s: %v", id, err)
			// Optionally remove dead streams here:
			// delete(s.clients, id)
		}
	}

	log.Printf("[BROADCAST] Type=%v | From=%s | Lamport=%d | Message=\"%s\"",
		b.Type, b.ClientId, b.LamportClock, b.Message)
}

// -----------------------------------------------------
// main: start gRPC server, add graceful shutdown
// -----------------------------------------------------
func main() {
	// Timestamps are included by default; add a component prefix.
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[SERVER] ")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv := &ChitChatServer{
		clients: make(map[string]pb.ChitChat_JoinServer),
		lamport: 0,
	}
	pb.RegisterChitChatServer(grpcServer, srv)

	// Handle Ctrl+C / SIGTERM for graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		log.Println("[SERVER] Shutdown signal received, stopping...")
		grpcServer.GracefulStop()
	}()

	log.Println("[SERVER] Started on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
