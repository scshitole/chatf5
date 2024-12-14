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

	// For testing, first process test commands to verify functionality
	log.Println("Executing test commands...")
	
	// Test Virtual Servers
	log.Println("Testing Virtual Servers listing...")
	vsResponse, err := chatInterface.ProcessQuery("show virtual servers")
	if err != nil {
		log.Printf("Error with virtual servers test: %v\n", err)
	} else {
		log.Printf("Virtual servers test successful, found servers in response")
		fmt.Printf("\nBIG-IP Virtual Servers: %s\n", vsResponse)
	}

	// Test WAF Policies with Virtual Server Associations
	log.Println("\n=== Starting WAF Policy and Virtual Server Association Test ===")
	log.Println("Step 1: Testing WAF/ASM module availability...")
	testQueries := []string{
		"list the WAF policy and the virtual server on which the policy is applied",
		"show WAF policies with their virtual servers",
		"display all WAF policy to virtual server mappings",
	}
	
	for _, query := range testQueries {
		log.Printf("\nTesting query: %s", query)
		wafResponse, err := chatInterface.ProcessQuery(query)
		if err != nil {
			log.Printf("Error with WAF policies test: %v\n", err)
			log.Printf("WAF Error Details: %+v\n", err)
			log.Printf("\nTroubleshooting Steps:")
			log.Printf("1. Verify ASM module is provisioned")
			log.Printf("2. Check user permissions for ASM policy access")
			log.Printf("3. Confirm BIG-IP version supports ASM/WAF features")
			log.Printf("4. Verify virtual server associations are accessible")
			continue
		}
		
		log.Printf("WAF policies and virtual server associations test completed successfully")
		fmt.Printf("\nBIG-IP WAF Policies and Their Virtual Server Associations:\n%s\n", wafResponse)
		
		// On successful query, test specific policy details
		if strings.Contains(wafResponse, "VS_WAF") {
			log.Printf("\nStep 2: Testing specific WAF policy details with virtual server bindings...")
			detailResponse, detailErr := chatInterface.ProcessQuery("show policy details VS_WAF")
			if detailErr != nil {
				log.Printf("Note: Could not fetch detailed policy information: %v", detailErr)
			} else {
				log.Printf("Successfully retrieved WAF policy details with virtual server bindings")
				fmt.Printf("\nDetailed Policy Information:\n%s\n", detailResponse)
			}
			break // Exit after successful test
		}
	}
	log.Println("=== WAF Policy and Virtual Server Association Test Complete ===\n")
	

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
