package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Message struct {
	From string
	Seq  int
	Ack  int
	Flag string
}

func client(client2server chan<- Message, server2client <-chan Message) {
	// random starting seq number x
	x := rand.Intn(1000)

	// Step 1: send SYN
	syn := Message{From: "Client", Seq: x, Ack: 0, Flag: "SYN"}
	fmt.Printf("Client -> Server  SYN  seq=%d\n", x)
	client2server <- syn
	fmt.Println("Client: SYN-SENT")

	// Step 2: wait for SYN-ACK from server
	reply := <-server2client
	fmt.Printf("Client <- Server  SYN-ACK  seq=%d ack=%d\n", reply.Seq, reply.Ack)

	// Verify SYN-ACK
	if reply.Flag != "SYN-ACK" || reply.Ack != x+1 {
		fmt.Println("Client: invalid response, handshake failed")
		return
	}

	// Step 3: send final ACK
	ack := Message{From: "Client", Seq: x + 1, Ack: reply.Seq + 1, Flag: "ACK"}
	fmt.Printf("Client -> Server  ACK  seq=%d ack=%d\n", ack.Seq, ack.Ack)
	client2server <- ack
	fmt.Println("Client: Connection ESTABLISHED!")
}

func server(client2server <-chan Message, server2client chan<- Message) {
	fmt.Println("Server: LISTEN")

	// Step 1: wait for SYN from client
	syn := <-client2server
	fmt.Printf("Server <- Client  SYN  seq=%d\n", syn.Seq)

	// Verify SYN
	if syn.Flag != "SYN" {
		fmt.Println("Server: invalid SYN, handshake failed")
		return
	}

	x := syn.Seq
	y := rand.Intn(1000)

	// Step 2: send SYN-ACK
	reply := Message{From: "Server", Seq: y, Ack: x + 1, Flag: "SYN-ACK"}
	fmt.Println("Server: SYN-RECEIVED")
	fmt.Printf("Server -> Client  SYN-ACK  seq=%d ack=%d\n", y, x+1)
	server2client <- reply

	// Step 3: wait for final ACK
	ack := <-client2server
	fmt.Printf("Server <- Client  ACK  seq=%d ack=%d\n", ack.Seq, ack.Ack)

	// Verify ACK
	if ack.Flag != "ACK" || ack.Ack != y+1 {
		fmt.Println("Server: invalid ACK, handshake failed")
		return
	}

	fmt.Println("Server: Handshake connection established")
}

func main() {

	client2server := make(chan Message)
	server2client := make(chan Message)

	go client(client2server, server2client)
	go server(client2server, server2client)

	time.Sleep(time.Second)
}
