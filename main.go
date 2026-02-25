package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"f5chat/bigip"
	"f5chat/chat"
	"f5chat/config"
	"f5chat/llm"
)

func runSelfTest(chatInterface *chat.Interface) error {
	testQueries := []string{
		"show virtual servers",
		"list the WAF policy and the virtual server on which the policy is applied",
		"show WAF policies with their virtual servers",
		"display all WAF policy to virtual server mappings",
	}

	for _, query := range testQueries {
		response, err := chatInterface.ProcessQuery(query)
		if err != nil {
			fmt.Printf("SELF-TEST FAIL: %q -> %v\n", query, err)
			return err
		}

		if strings.TrimSpace(response) == "" {
			err = fmt.Errorf("empty response")
			fmt.Printf("SELF-TEST FAIL: %q -> %v\n", query, err)
			return err
		}

		fmt.Printf("SELF-TEST PASS: %q\n", query)
	}

	return nil
}

func main() {
	selfTestFlag := flag.Bool("self-test", false, "run startup self-test queries and exit")
	flag.Parse()

	selfTestEnabled := *selfTestFlag || strings.EqualFold(os.Getenv("SELF_TEST"), "true")

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
	if selfTestEnabled {
		if err := runSelfTest(chatInterface); err != nil {
			os.Exit(1)
		}

		fmt.Println("SELF-TEST PASS")
		os.Exit(0)
	}

	fmt.Println("Welcome to F5 BIG-IP Chat Interface!")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("----------------------------------------")

	reader := bufio.NewReader(os.Stdin)

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
