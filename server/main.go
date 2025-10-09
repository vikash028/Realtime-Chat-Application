package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Client represents a connected client
type Client struct {
	conn   net.Conn
	writer *bufio.Writer
}

// Hub manages all connected clients
type Hub struct {
	clients map[*Client]bool
	mu      sync.Mutex
}

// Global hub instance
var hub = Hub{
	clients: make(map[*Client]bool),
}

// removeClient removes a client from the hub
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client)
	fmt.Printf("Total clients: %d\n", len(h.clients))
}

// addClient adds a client to the hub
func (h *Hub) addClient(client *Client) {
	//prevents race condition (when two users enters at same)
	h.mu.Lock()         //lock when a user enters
	defer h.mu.Unlock() //unlock for other users after function ends

	h.clients[client] = true
	fmt.Printf("Total clients: %d\n", len(h.clients))
}

// Broadcast sends a message to all clients EXCEPT sender
func (h *Hub) broadcast(message string, sender *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		//skip the sender
		if client == sender {
			continue
		}

		//send message to this client
		_, err := client.writer.WriteString(message)
		if err != nil {
			log.Println("Error writing to client:", err)
			continue
		}

		//Flush the buffer to ensure message is sent immediately
		err = client.writer.Flush()
		if err != nil {
			log.Println("Error flushing to client:", err)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	defer fmt.Println("Client disconnected:", conn.RemoteAddr())

	//create client object
	client := &Client{
		conn:   conn,
		writer: bufio.NewWriter(conn),
	}

	//Add client to hub
	hub.addClient(client)
	//Remove client from the hub after the function ends
	defer hub.removeClient(client)

	//Read messages from this client
	reader := bufio.NewReader(conn)

	for {
		//Read message from client
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading:", err)
			}
			break
		}

		fmt.Printf("Recieved: %s", message)

		//Broadcast to all clients EXCEPT sender
		hub.broadcast(message, client)
	}
}

func main() {
	//Listen on TCP port 7007
	listener, err := net.Listen("tcp", ":7007")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("Chat server listening on port 7007...")

	//Accept connections continuously
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		//Prints the IP address of the client
		fmt.Println("Client connected:", conn.RemoteAddr())

		//Handle each client in a seperate goroutine (non-blocking!)
		go handleClient(conn)
	}
}
