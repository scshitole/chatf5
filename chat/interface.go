package chat

import (
	"fmt"
	"log"
	"strings"

	"f5chat/bigip"
	"f5chat/llm"
	"f5chat/utils"
)

type Interface struct {
	bigipClient *bigip.Client
	llmClient   *llm.OpenAIClient
}

func NewInterface(bigipClient *bigip.Client, llmClient *llm.OpenAIClient) *Interface {
	return &Interface{
		bigipClient: bigipClient,
		llmClient:   llmClient,
	}
}

func (i *Interface) ProcessQuery(query string) (string, error) {
	// First, use LLM to understand the intent and get structured response
	llmResponse, err := i.llmClient.ProcessPrompt(query)
	if err != nil {
		return "", fmt.Errorf("I apologize, but I'm having trouble understanding your request. Could you please rephrase it? (Error: %v)", err)
	}

	// Execute the appropriate BIG-IP operation based on LLM response
	response, err := i.executeOperation(llmResponse, query)
	if err != nil {
		return "", fmt.Errorf("I understood your request about the BIG-IP configuration, but encountered an issue while fetching the information. Please try again. (Error: %v)", err)
	}

	return response, nil
}

// containsAny checks if the text contains any of the given phrases
func containsAny(text string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func (i *Interface) executeOperation(llmResponse string, originalQuery string) (string, error) {
	// Enhanced intent detection with common variations
	lowerResponse := strings.ToLower(llmResponse)

	// WAF Policy related queries
	if containsAny(lowerResponse, []string{
		"waf policy", "waf policies", 
		"web application firewall", 
		"asm policy", "asm policies",
		"application security",
		"application firewall",
		"security policy",
	}) {
		log.Printf("Detected WAF policy query request: %s", originalQuery)
		
		// Check if looking for a specific policy
		if strings.Contains(lowerResponse, "details") && strings.Contains(lowerResponse, "policy") {
			// Extract policy name from the query
			words := strings.Fields(lowerResponse)
			var policyName string
			for i, word := range words {
				if word == "policy" && i+1 < len(words) {
					policyName = words[i+1]
					break
				}
			}
			
			if policyName != "" {
				policy, err := i.bigipClient.GetWAFPolicyDetails(policyName)
				if err != nil {
					log.Printf("Error fetching WAF policy details: %v", err)
					return "", fmt.Errorf("failed to fetch WAF policy details: %v", err)
				}
				log.Printf("Successfully retrieved WAF policy details for %s", policyName)
				return utils.FormatWAFPolicyDetails(policy), nil
			}
		}
		
		// Default: list all policies
		policies, err := i.bigipClient.GetWAFPolicies()
		if err != nil {
			log.Printf("Error fetching WAF policies: %v", err)
			return "", fmt.Errorf("failed to fetch WAF policies: %v", err)
		}
		log.Printf("Successfully retrieved WAF policies")
		return utils.FormatWAFPolicies(policies), nil
	}

	// Virtual Server related queries
	if containsAny(lowerResponse, []string{"virtual server", "vip", "virtual ip", "virtual address"}) {
		vs, err := i.bigipClient.GetVirtualServers()
		if err != nil {
			return "", err
		}
		return utils.FormatVirtualServers(vs), nil
	}

	// Pool related queries
	if containsAny(lowerResponse, []string{"pool", "server pool", "backend pool", "server group"}) {
		pools, poolMembers, err := i.bigipClient.GetPools()
		if err != nil {
			return "", err
		}
		return utils.FormatPools(pools, poolMembers), nil
	}

	// Node related queries
	if containsAny(lowerResponse, []string{"node", "server", "backend", "real server"}) {
		nodes, err := i.bigipClient.GetNodes()
		if err != nil {
			return "", err
		}
		return utils.FormatNodes(nodes), nil
	}

	return "I understand you're asking about BIG-IP configuration. To help you better, could you please be more specific?\n\n" +
		"You can ask questions like:\n" +
		"1. 'Show me all virtual servers (VIPs)' - View front-end service points\n" +
		"2. 'List all pools and their members' - See load balancing groups\n" +
		"3. 'Display node status' - Check backend server health\n\n" +
		"Feel free to ask about specific components or use natural language to describe what you're looking for.", nil
}