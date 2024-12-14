package bigip

import (
	"crypto/tls"
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

	// Set custom transport with TLS configuration matching successful curl command
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Required for self-signed cert
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS12, // Force TLSv1.2 as seen in curl
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Match curl's cipher
			},
		},
		// Timeouts and settings matching curl's successful connection
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       10 * time.Second,
		DisableCompression:    true,
		DisableKeepAlives:     true, // Match curl's behavior
		ForceAttemptHTTP2:     false, // Force HTTP/1.1 as seen in curl
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		// Force direct connection without proxy
		Proxy: nil,
	}

	log.Printf("Configuring TLS transport with custom settings...")
	bigipClient.Transport = customTransport

	// Test connection with timeout using a simple endpoint
	log.Printf("Starting connection test to BIG-IP at %s", host)
	log.Printf("Full connection URL: %s/mgmt/tm/ltm/virtual", baseURL)
	log.Printf("Network config:")
	log.Printf("- Direct connection (no proxy)")
	log.Printf("- Allow reused connections: false")
	log.Printf("- Keep-alive enabled: false")
	log.Printf("- HTTP/1.1 enforced (HTTP/2 disabled)")
	log.Printf("- Using TLSv1.2 with ECDHE-RSA-AES128-GCM-SHA256")
	log.Printf("- Self-signed certificate handling: enabled")
	log.Printf("Connection Details:")
	log.Printf("- Protocol: HTTPS/TLS")
	log.Printf("- TLS Version: 1.2 (forced)")
	log.Printf("- Cipher Suite: ECDHE-RSA-AES128-GCM-SHA256")
	log.Printf("- Certificate Verification: Disabled (allowing self-signed)")
	log.Printf("- Auth Method: Basic Auth")
	log.Printf("- Username: %s (length: %d)", cfg.BigIPUsername, len(cfg.BigIPUsername))
	log.Printf("- Password length: %d characters", len(cfg.BigIPPassword))

	// Additional debug info
	log.Printf("Network settings:")
	log.Printf("1. Port: %s (management port)", port)
	log.Printf("2. TLS config: TLSv1.2 only")
	log.Printf("3. Cipher: ECDHE-RSA-AES128-GCM-SHA256")
	log.Printf("4. Keep-alive: enabled")
	log.Printf("5. Connection timeout: %v", customTransport.TLSHandshakeTimeout)

	// Validate credentials
	if cfg.BigIPUsername != "admin" || len(cfg.BigIPPassword) != 10 {
		log.Printf("WARNING: Credential validation failed:")
		log.Printf("- Expected username: 'admin', Got: '%s'", cfg.BigIPUsername)
		log.Printf("- Expected password length: 10, Got: %d", len(cfg.BigIPPassword))
	}

	// Detailed TLS debug logging
	log.Printf("TLS Configuration Details:")
	log.Printf("- Min TLS Version: 1.2")
	log.Printf("- Max TLS Version: 1.2")
	log.Printf("- Allowed Cipher Suite: ECDHE-RSA-AES128-GCM-SHA256")
	log.Printf("- Certificate Verification: Disabled for development")
	log.Printf("- Target URL: %s", baseURL)

	done := make(chan error, 1)
	go func() {
		// Try a lightweight API call to test connectivity
		log.Println("Attempting to fetch VirtualServers as connection test...")
		vs, err := bigipClient.VirtualServers()
		if err != nil {
			log.Printf("Connection test API call failed with detailed error: %+v", err)
			log.Printf("Error type: %T", err)
			errLower := strings.ToLower(err.Error())

			// Enhanced error classification with specific details
			switch {
			case strings.Contains(errLower, "certificate") || strings.Contains(errLower, "x509"):
				log.Printf("TLS/Certificate error: The server's certificate could not be verified")
				log.Printf("This is expected for self-signed certificates in development")
				// Continue despite certificate error as we're using InsecureSkipVerify
				vs, err = bigipClient.VirtualServers()
				if err == nil {
					log.Printf("Connection succeeded after ignoring certificate error")
					done <- nil
					return
				}
			case strings.Contains(errLower, "connection refused"):
				log.Printf("Connection refused - double checking settings:")
				log.Printf("1. Target: %s:%s (matching curl)", host, port)
				log.Printf("2. TLS Version: 1.2 (forced)")
				log.Printf("3. Cipher: ECDHE-RSA-AES128-GCM-SHA256")
				log.Printf("4. Direct connection without proxy")
			case strings.Contains(errLower, "timeout"):
				log.Printf("Connection timeout - verifying network:")
				log.Printf("1. URL: %s/mgmt/tm/ltm/virtual", baseURL)
				log.Printf("2. Timeouts: TLS=%ds, Response=%ds",
					int(customTransport.TLSHandshakeTimeout.Seconds()),
					int(customTransport.ResponseHeaderTimeout.Seconds()))
			case strings.Contains(errLower, "unauthorized") ||
				strings.Contains(errLower, "authentication") ||
				strings.Contains(errLower, "401"):
				log.Printf("Authentication failed - verifying:")
				log.Printf("1. Username: %s (expected: admin)", cfg.BigIPUsername)
				log.Printf("2. Password length: %d (expected: 10)", len(cfg.BigIPPassword))
			default:
				log.Printf("Unexpected error type: %v", err)
				log.Printf("Full error context: %#v", err)
			}

			done <- fmt.Errorf("connection test failed: %v", err)
			return
		}

		// Log successful connection
		log.Printf("Successfully connected to BIG-IP!")
		log.Printf("Found %d virtual servers", len(vs.VirtualServers))
		log.Printf("Connection test passed - API responding correctly")
		done <- nil
	}()

	// Wait for connection test with timeout
	select {
	case err := <-done:
		if err != nil {
			log.Printf("Connection test failed: %v", err)
			return nil, fmt.Errorf("failed to connect to BIG-IP: %v", err)
		}
		log.Println("Successfully connected to BIG-IP")
	case <-time.After(30 * time.Second):
		log.Printf("Connection attempt timed out after 30 seconds")
		log.Printf("Detailed Connection Information:")
		log.Printf("- Target Host: %s", host)
		log.Printf("- Management Port: 8443")
		log.Printf("- TLS Version: 1.2/1.3")
		log.Printf("- Certificate Verification: Disabled (allowing self-signed)")
		log.Printf("- Connection Timeouts:")
		log.Printf("  * TLS Handshake: 10s")
		log.Printf("  * Response Header: 10s")
		log.Printf("  * Total Connection: 30s")
		return nil, fmt.Errorf("connection timeout - since browser access works, this might be a TLS handshake issue. Please verify:\n1. HTTPS/TLS connectivity\n2. Certificate handling\n3. BIG-IP API endpoint status")
	}

	return &Client{bigipClient}, nil
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