package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"f5chat/bigip"
	"f5chat/chat"
	"f5chat/config"
	"f5chat/llm"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Attempting to connect to BIG-IP...")
	bigipClient, err := bigip.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize BIG-IP client: %v", err)
	}
	log.Println("Successfully connected to BIG-IP")

	log.Println("Initializing OpenAI client...")
	llmClient, err := llm.NewOpenAIClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize OpenAI client: %v", err)
	}
	log.Println("OpenAI client initialized successfully")

	// Initialize chat interface
	chatInterface := chat.NewInterface(bigipClient, llmClient)

	fmt.Println("Welcome to F5 BIG-IP Chat Interface!")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("----------------------------------------")

	reader := bufio.NewReader(os.Stdin)

	// For testing, first process a test command to list virtual servers
	log.Println("Executing test command: show virtual servers")
	testResponse, err := chatInterface.ProcessQuery("show virtual servers")
	if err != nil {
		log.Printf("Error with test command: %v\n", err)
		fmt.Printf("Error with test command: %v\n", err)
	} else {
		log.Printf("Test command successful, response length: %d", len(testResponse))
		fmt.Printf("\nBIG-IP: %s\n", testResponse)
	}

	// Then continue with the normal interactive loop
	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}

		response, err := chatInterface.ProcessQuery(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nBIG-IP: %s\n", response)
	}
}
