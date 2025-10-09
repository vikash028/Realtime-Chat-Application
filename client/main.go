package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// receiveMessages listens for messages from the server
func receiveMessages(conn net.Conn, done chan bool) {

	serverReader := bufio.NewReader(conn)

	for {
		//Read messages from server
		message, err := serverReader.ReadString('\n')
		if err != nil {
			//server disconnected or error
			done <- true
			return
		}

		//Print what server echoed back
		fmt.Print(message)
	}
}

// sendMessages handles user input and sends to server
func sendMessages(conn net.Conn, username string, reader *bufio.Reader, done chan bool) {

	for {
		//check if server disconnected
		select {
		case <-done:
			fmt.Println("\nDisconnected from server")
			return
		default:
			//continue to read input
		}

		//read from user input
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			return
		}

		//check if user wants to quit
		trimmed := strings.TrimSpace(message)
		if trimmed == "quit" || trimmed == "exit" {
			fmt.Println("Disconnecting...")
			return
		}

		//Format message with username
		formattedMessage := username + ": " + message

		//Echo locally (show what user typed)
		fmt.Print(formattedMessage)

		//send to server
		_, err = conn.Write([]byte(formattedMessage))
		if err != nil {
			log.Println("Error sending message:", err)
			return
		}
	}
}

func main() {

	//Ask for username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("name: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading username:", err)
	}
	username = strings.TrimSpace(username)

	//connect to the echo server
	conn, err := net.Dial("tcp", "localhost:7007")
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()

	fmt.Println("Welcome to the chatroom!")

	//creates a channel to signal when to exit
	done := make(chan bool)

	//Goroutine to receive messages from server
	go receiveMessages(conn, done)

	//Main goroutine handles user input
	sendMessages(conn, username, reader, done)
}
