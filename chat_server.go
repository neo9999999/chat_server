package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn     net.Conn
	username string
}

var (
	clients = make(map[net.Conn]Client)
	mu      sync.Mutex
)

func broadcastMessage(message string, sender net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for conn := range clients {
		if conn != sender {
			fmt.Fprintln(conn, message)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	fmt.Fprint(conn, "Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	client := Client{conn: conn, username: username}

	mu.Lock()
	clients[conn] = client
	mu.Unlock()

	welcomeMessage := fmt.Sprintf("** %s has joined the chat **", client.username)
	broadcastMessage(welcomeMessage, conn)
	fmt.Fprintln(conn, "Welcome to the chat room!")

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		if message == "/exit" {
			disconnectMessage := fmt.Sprintf("** %s has left the chat **", client.username)
			broadcastMessage(disconnectMessage, conn)
			break
		}

		chatMessage := fmt.Sprintf("[%s]: %s", client.username, message)
		broadcastMessage(chatMessage, conn)
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}

func main() {
	fmt.Println("Starting chat server on :8080...")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleClient(conn)
	}
}
