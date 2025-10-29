Chit Chat is a distributed chat app where users can join, chat, and leave whenever they want.
It’s built in Go using gRPC and Protocol Buffers, and showcases core distributed systems concepts like communication, coordination, concurrency, and logical time (Lamport timestamps).
-------------------------------------------------------------------------------------
Structure:
project-root/
├── client/ # Client code
├── server/ # Server code
├── grpc/ # .proto file
└── readme.md
-------------------------------------------------------------------------------------
Instructions: 
Turn on server:
go to DistributedSystemsGroup directonary
1. "cd server"
2. "go run main.go" 
3  "control c" to turn off the server

Add clients:
go to DistributedSystemsGroup directonary
1. "cd client"
2. "go run main.go"

Console will ask for name, write an name and press enter
to write message, simply write in console and press enter
to leave the chat write /exit
-------------------------------------------------------------------------------------
Example: 

Server - [SERVER] 2025/10/29 18:47:58 [SERVER] Started on port 50051
[SERVER] 2025/10/29 18:48:12 [EVENT] Connection | User=hedin
[SERVER] 2025/10/29 18:48:12 [EVENT] Join | User=hedin | Lamport=1
[SERVER] 2025/10/29 18:48:12 [BROADCAST] Type=JOIN | From=hedin | Lamport=1 | Message="hedin joined the chat"
[SERVER] 2025/10/29 18:48:23 [EVENT] Connection | User=oscar
[SERVER] 2025/10/29 18:48:23 [EVENT] Join | User=oscar | Lamport=2
[SERVER] 2025/10/29 18:48:23 [BROADCAST] Type=JOIN | From=oscar | Lamport=2 | Message="oscar joined the chat"
[SERVER] 2025/10/29 18:48:27 [EVENT] Message | User=hedin | Lamport=3 | Text="hello"
[SERVER] 2025/10/29 18:48:27 [BROADCAST] Type=CHAT | From=hedin | Lamport=3 | Message="hello"

Client - > 2025/10/29 18:48:12 [EVENT] Join | User=hedin | Lamport=1
2025/10/29 18:48:23 [EVENT] Join | User=oscar | Lamport=2
hello
2025/10/29 18:48:27 [EVENT] Message | User=hedin | Lamport=3 | Text="hello"




