package bigip

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/f5devcentral/go-bigip"
	"f5chat/config"
)

type Client struct {
	*bigip.BigIP
	Username string
	Password string
}

func NewClient(cfg *config.Config) (*Client, error) {
	log.Printf("Raw BIG-IP host from environment: %s", cfg.BigIPHost)

	// Parse host and port
	hostParts := strings.Split(strings.TrimSpace(cfg.BigIPHost), ":")
	host := hostParts[0]
	port := "443" // default HTTPS port
	if len(hostParts) > 1 {
		port = hostParts[1]
	}

	log.Printf("Parsed host components - Host: %s, Port: %s", host, port)

	// Construct proper URL
	baseURL := fmt.Sprintf("https://%s:%s", host, port)
	log.Printf("Constructed base URL: %s", baseURL)

	// Create configuration for BIG-IP session
	config := &bigip.Config{
		Address:  baseURL,
		Username: cfg.BigIPUsername,
		Password: cfg.BigIPPassword,
	}

	log.Printf("Creating BIG-IP session with configuration: Address=%s, Username=%s",
		config.Address, config.Username)

	bigipClient := bigip.NewSession(config)
	log.Printf("BIG-IP session created, attempting API connection...")

	// Set custom transport with enhanced TLS configuration for HTTPS
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Required for self-signed certificates
			MinVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		},
		// Enhanced connection settings
		TLSHandshakeTimeout:   45 * time.Second,
		ResponseHeaderTimeout: 45 * time.Second,
		ExpectContinueTimeout: 15 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		ForceAttemptHTTP2:     false, // Stick to HTTP/1.1 for better compatibility
	}

	log.Printf("Configuring TLS transport with custom settings...")
	bigipClient.Transport = customTransport

	// Test connection with timeout using a simple endpoint
	log.Printf("Starting connection test to BIG-IP at %s", host)
	log.Printf("Using HTTPS connection to %s/mgmt/tm/ltm/virtual", baseURL)
	log.Printf("Connection details:")
	log.Printf("- Protocol: HTTPS")
	log.Printf("- Host: %s", host)
	log.Printf("- Port: %s", port)
	log.Printf("- Username: %s", cfg.BigIPUsername)
	log.Printf("- TLS: Enabled with InsecureSkipVerify")
	log.Printf("- Timeout settings:")
	log.Printf("  * TLS Handshake: 45s")
	log.Printf("  * Response Header: 45s")
	log.Printf("  * Connection Idle: 90s")

	// Create a channel for connection result
	connectionStatus := make(chan error, 1)

	// Maximum number of retries
	maxRetries := 3
	retryDelay := 5 * time.Second

	// Start connection test in a goroutine
	go func() {
		var lastErr error
		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				log.Printf("Retry attempt %d/%d after %v delay...", retry+1, maxRetries, retryDelay)
				time.Sleep(retryDelay)
			}

			// Try to fetch virtual servers as a connection test
			testVs, testErr := bigipClient.VirtualServers()
			if testErr == nil {
				log.Printf("Connection successful on attempt %d, found %d virtual servers", retry+1, len(testVs.VirtualServers))
				connectionStatus <- nil
				return
			}

			lastErr = testErr
			errLower := strings.ToLower(testErr.Error())
			log.Printf("Connection attempt %d failed: %v", retry+1, testErr)

			// Handle different error cases
			switch {
			case strings.Contains(errLower, "certificate"):
				log.Printf("Certificate validation error - modifying TLS config and retrying...")
				bigipClient.Transport = customTransport
				// Immediate retry with new transport
				retryVs, retryErr := bigipClient.VirtualServers()
				if retryErr == nil {
					log.Printf("Connection successful after certificate handling, found %d virtual servers", len(retryVs.VirtualServers))
					connectionStatus <- nil
					return
				}
				log.Printf("Still failed after certificate handling: %v", retryErr)
				
			case strings.Contains(errLower, "connection refused"):
				log.Printf("Connection refused - port %s might be blocked or BIG-IP not accepting connections", port)
				log.Printf("Please verify:\n1. BIG-IP management port is accessible\n2. No firewall rules blocking port %s", port)
				
			case strings.Contains(errLower, "no such host"):
				log.Printf("DNS resolution failed for host: %s\nPlease verify the hostname/IP is correct", host)
				
			case strings.Contains(errLower, "timeout"):
				log.Printf("Connection timed out - possible network issues:\n1. Slow network connection\n2. Firewall blocking traffic\n3. BIG-IP high load or not responding")
				
			case strings.Contains(errLower, "unauthorized"):
				log.Printf("Authentication failed - verify credentials:\n1. Username (current: %s)\n2. Password (length: %d)\n3. BIG-IP user permissions", 
					cfg.BigIPUsername, len(cfg.BigIPPassword))
				
			default:
				log.Printf("Unexpected error: %v\nFull error details: %#v", testErr, testErr)
			}
		}

		// If we get here, all retries failed
		connectionStatus <- fmt.Errorf("failed to connect after %d attempts - last error: %v", maxRetries, lastErr)
	}()

	// Wait for connection test with timeout
	select {
	case err := <-connectionStatus:
		if err != nil {
			log.Printf("All connection attempts failed: %v", err)
			return nil, fmt.Errorf("failed to connect to BIG-IP: %v", err)
		}
		log.Printf("Successfully connected to BIG-IP")
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("connection timeout after 30 seconds - please verify:\n1. BIG-IP host and port (%s)\n2. Network connectivity\n3. Firewall rules", cfg.BigIPHost)
	}

	// The first connection test was successful, no need for a second test
	log.Println("Connection to BIG-IP established successfully")

	return &Client{
		BigIP:    bigipClient,
		Username: cfg.BigIPUsername,
		Password: cfg.BigIPPassword,
	}, nil
}

func (c *Client) GetVirtualServers() ([]bigip.VirtualServer, error) {
	log.Println("\n=== Starting GetVirtualServers Operation ===")
	log.Printf("Endpoint: /mgmt/tm/ltm/virtual")
	log.Printf("Method: GET")
	log.Printf("Authentication: Basic Auth (Username: %s)", c.User)

	// Make the API request with detailed logging
	log.Println("\nMaking API request to fetch virtual servers...")
	vs, err := c.VirtualServers()
	if err != nil {
		log.Printf("\nERROR: Failed to fetch virtual servers")
		log.Printf("Error Type: %T", err)
		log.Printf("Error Message: %v", err)

		// Enhanced error analysis
		errStr := err.Error()
		switch {
		case strings.Contains(strings.ToLower(errStr), "unauthorized"):
			log.Printf("Authentication Error: Please verify credentials")
			log.Printf("Expected: Username='admin', Password length=10")
		case strings.Contains(strings.ToLower(errStr), "connection"):
			log.Printf("Connection Error: Unable to reach BIG-IP")
			log.Printf("Please verify network connectivity and firewall rules")
		case strings.Contains(strings.ToLower(errStr), "certificate"):
			log.Printf("TLS Certificate Error: Certificate validation failed")
			log.Printf("This is expected for self-signed certificates")
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
		}
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	log.Println("\nAPI Response received successfully")

	// Process the response with detailed logging
	var virtualServers []bigip.VirtualServer
	if vs != nil && vs.VirtualServers != nil {
		count := len(vs.VirtualServers)
		log.Printf("\nFound %d virtual server(s)", count)

		for i, v := range vs.VirtualServers {
			log.Printf("\nVirtual Server [%d/%d]:", i+1, count)
			log.Printf("  Name:        %s", v.Name)
			log.Printf("  Destination: %s", v.Destination)
			log.Printf("  Pool:        %s", v.Pool)
			log.Printf("  Status:      %s", map[bool]string{true: "Enabled", false: "Disabled"}[v.Enabled])
			virtualServers = append(virtualServers, v)
		}
	} else {
		log.Printf("\nWARNING: No virtual servers found")
		log.Printf("Response validation:")
		log.Printf("- vs object is nil: %v", vs == nil)
		log.Printf("- vs.VirtualServers is nil: %v", vs != nil && vs.VirtualServers == nil)
	}

	log.Printf("GetVirtualServers operation completed. Returning %d virtual servers", len(virtualServers))
	return virtualServers, nil
}

func (c *Client) GetPools() ([]bigip.Pool, map[string][]string, error) {
	pools, err := c.Pools()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pools: %v", err)
	}

	// Convert *Pools to []Pool
	var poolList []bigip.Pool
	poolMembers := make(map[string][]string)

	for _, p := range pools.Pools {
		poolList = append(poolList, p)
		// Get members for each pool
		members, err := c.PoolMembers(p.Name)
		if err != nil {
			fmt.Printf("Warning: failed to get members for pool %s: %v\n", p.Name, err)
			continue
		}
		var memberList []string
		if members != nil {
			for i := range members.PoolMembers {
				memberList = append(memberList, members.PoolMembers[i].FullPath)
			}
		}
		poolMembers[p.Name] = memberList
	}
	return poolList, poolMembers, nil
}

func (c *Client) GetNodes() ([]bigip.Node, error) {
	nodes, err := c.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	// Convert *Nodes to []Node
	var nodeList []bigip.Node
	for _, n := range nodes.Nodes {
		nodeList = append(nodeList, n)
	}
	return nodeList, nil
}
// ASMPolicy represents a WAF/ASM policy in BIG-IP
type ASMPolicy struct {
	Name          string `json:"name"`
	FullPath      string `json:"fullPath"`
	ID            string `json:"id"`
	Description   string `json:"description,omitempty"`
	Active        bool   `json:"active"`
	Type          string `json:"type,omitempty"`
	EnforcementMode string `json:"enforcementMode,omitempty"`
	Kind          string `json:"kind,omitempty"`
	SelfLink      string `json:"selfLink,omitempty"`
}

// ASMPoliciesResponse represents the response from BIG-IP for ASM policies
type ASMPoliciesResponse struct {
	Items []ASMPolicy `json:"items"`
	Kind  string     `json:"kind"`
	Generation int64 `json:"generation"`
	SelfLink string  `json:"selfLink"`
}

// GetWAFPolicies retrieves the list of WAF policies from BIG-IP
func (c *Client) GetWAFPolicies() ([]string, error) {
	log.Printf("\n=== Starting GetWAFPolicies Operation ===")
	log.Printf("Endpoint: /mgmt/tm/asm/policies")
	log.Printf("Method: GET")
	log.Printf("Authentication: Basic Auth (Username: %s)", c.Username)

	var policies ASMPoliciesResponse
	req := &bigip.APIRequest{
		Method:      "GET",
		URL:         "mgmt/tm/asm/policies",
		ContentType: "application/json",
	}
	resp, err := c.BigIP.APICall(req)
	if err != nil {
		log.Printf("\nERROR: Failed to fetch WAF policies")
		log.Printf("Error Type: %T", err)
		log.Printf("Error Message: %v", err)
		return nil, fmt.Errorf("failed to get WAF policies: %v", err)
	}
	err = json.Unmarshal(resp, &policies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WAF policies response: %v", err)
	}

	log.Printf("\nAPI Response received successfully")
	log.Printf("Response Kind: %s", policies.Kind)
	log.Printf("Generation: %d", policies.Generation)

	var policyNames []string
	for _, policy := range policies.Items {
		log.Printf("\nProcessing policy:")
		log.Printf("  Name: %s", policy.Name)
		log.Printf("  ID: %s", policy.ID)
		log.Printf("  Type: %s", policy.Type)
		log.Printf("  Enforcement Mode: %s", policy.EnforcementMode)
		policyNames = append(policyNames, policy.Name)
	}

	log.Printf("\nFound %d WAF policies", len(policyNames))
	return policyNames, nil
}