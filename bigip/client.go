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

// Client wraps the F5 BIG-IP client with additional functionality
type Client struct {
	*bigip.BigIP
	Username string
	Password string
}

// VirtualServer represents a BIG-IP virtual server configuration
type VirtualServer struct {
	*bigip.VirtualServer
}

// Pool represents a BIG-IP server pool configuration
type Pool struct {
	*bigip.Pool
}

// Node represents a BIG-IP backend node configuration
type Node struct {
	*bigip.Node
}

// WAFPolicy represents a BIG-IP WAF (ASM) policy
type WAFPolicy struct {
	Name             string                 `json:"name"`
	FullPath         string                 `json:"fullPath"`
	ID               string                 `json:"id"`
	Description      string                 `json:"description,omitempty"`
	Active           bool                   `json:"active"`
	Type             string                 `json:"type,omitempty"`
	EnforcementMode  string                 `json:"enforcementMode,omitempty"`
	Kind             string                 `json:"kind,omitempty"`
	SelfLink         string                 `json:"selfLink,omitempty"`
	SignatureStaging bool                   `json:"signatureStaging,omitempty"`
	VirtualServers   []string              `json:"virtualServers,omitempty"`
	SignatureSetings map[string]interface{} `json:"signatureSettings,omitempty"`
	BlockingMode     string                 `json:"blockingMode,omitempty"`
	PlaceSignatures  bool                   `json:"placeSignaturesInStaging,omitempty"`
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
		TLSHandshakeTimeout:   45 * time.Second,
		ResponseHeaderTimeout: 45 * time.Second,
		ExpectContinueTimeout: 15 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		ForceAttemptHTTP2:     false,
	}

	log.Printf("Configuring TLS transport with custom settings...")
	bigipClient.Transport = customTransport

	// Test connection with timeout
	log.Printf("Starting connection test to BIG-IP at %s", host)
	log.Printf("Using HTTPS connection to %s/mgmt/tm/ltm/virtual", baseURL)

	// Create a channel for connection result
	connectionStatus := make(chan error, 1)

	// Maximum number of retries
	maxRetries := 3
	baseDelay := 5 * time.Second
	maxDelay := 30 * time.Second

	// Start connection test in a goroutine
	go func() {
		var lastErr error
		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				// Calculate exponential backoff delay
				backoffMultiplier := uint(1) << uint(retry-1)
				delay := baseDelay * time.Duration(backoffMultiplier)
				if delay > maxDelay {
					delay = maxDelay
				}
				log.Printf("Retry attempt %d/%d after %v delay (exponential backoff)...", retry+1, maxRetries, delay)
				time.Sleep(delay)
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

			switch {
			case strings.Contains(errLower, "certificate"):
				log.Printf("Certificate validation error - modifying TLS config and retrying...")
				bigipClient.Transport = customTransport
				retryVs, retryErr := bigipClient.VirtualServers()
				if retryErr == nil {
					log.Printf("Connection successful after certificate handling, found %d virtual servers", len(retryVs.VirtualServers))
					connectionStatus <- nil
					return
				}
				log.Printf("Still failed after certificate handling: %v", retryErr)
			case strings.Contains(errLower, "connection refused"):
				log.Printf("Connection refused - port %s might be blocked or BIG-IP not accepting connections", port)
			case strings.Contains(errLower, "no such host"):
				log.Printf("DNS resolution failed for host: %s", host)
			case strings.Contains(errLower, "timeout"):
				log.Printf("Connection timed out - possible network issues or firewall blocking")
			case strings.Contains(errLower, "unauthorized"):
				log.Printf("Authentication failed - verify username and password")
			default:
				log.Printf("Unexpected error: %v", testErr)
			}
		}
		connectionStatus <- fmt.Errorf("failed to connect after %d attempts - last error: %v", maxRetries, lastErr)
	}()

	// Wait for connection test with timeout
	select {
	case err := <-connectionStatus:
		if err != nil {
			return nil, fmt.Errorf("failed to connect to BIG-IP: %v", err)
		}
		log.Printf("Successfully connected to BIG-IP")
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("connection timeout after 60 seconds - please verify:\n1. BIG-IP host and port (%s)\n2. Network connectivity\n3. Firewall rules\n4. BIG-IP management interface status", cfg.BigIPHost)
	}

	return &Client{
		BigIP:    bigipClient,
		Username: cfg.BigIPUsername,
		Password: cfg.BigIPPassword,
	}, nil
}

// ASMPolicy represents detailed WAF/ASM policy information in BIG-IP
type ASMPolicy struct {
	WAFPolicy
	WhitelistIPs      []string                 `json:"whitelistIps,omitempty"`
	BlacklistIPs      []string                 `json:"blacklistIps,omitempty"`
	ModificationTime  string                   `json:"modificationTime,omitempty"`
	TemplateType     string                   `json:"templateType,omitempty"`
	TemplateReference map[string]interface{}   `json:"templateReference,omitempty"`
	ManualLock       bool                     `json:"manualLock,omitempty"`
	Parameters       map[string]interface{}    `json:"parameters,omitempty"`
	Attributes       map[string]interface{}    `json:"attributes,omitempty"`
	HasParent        bool                     `json:"hasParent,omitempty"`
	Links            map[string]interface{}    `json:"links,omitempty"`
}

// ASMPoliciesResponse represents the response from BIG-IP for ASM policies
type ASMPoliciesResponse struct {
	Items      []ASMPolicy `json:"items"`
	Kind       string      `json:"kind"`
	Generation int64       `json:"generation"`
	SelfLink   string      `json:"selfLink"`
}

// GetWAFPolicies retrieves the list of WAF policies from BIG-IP
func (c *Client) GetWAFPolicies() ([]*WAFPolicy, error) {
	log.Printf("\n=== Starting GetWAFPolicies Operation ===")
	log.Printf("Endpoint: /mgmt/tm/asm/policies")
	log.Printf("Method: GET")
	log.Printf("Authentication: Basic Auth (Username: %s)", c.Username)

	maxRetries := 3
	baseDelay := 5 * time.Second
	maxDelay := 30 * time.Second
	var lastErr error
	var policies ASMPoliciesResponse

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			// Calculate exponential backoff delay
			backoffMultiplier := uint(1) << uint(retry-1)
			delay := baseDelay * time.Duration(backoffMultiplier)
			if delay > maxDelay {
				delay = maxDelay
			}
			log.Printf("Retry attempt %d/%d for WAF policies after %v delay (exponential backoff)...", retry+1, maxRetries, delay)
			time.Sleep(delay)
		}

		req := &bigip.APIRequest{
			Method:      "GET",
			URL:         "mgmt/tm/asm/policies",
			ContentType: "application/json",
		}

		log.Printf("\nMaking API request to fetch WAF policies...")
		resp, err := c.BigIP.APICall(req)

		if err == nil {
			if err = json.Unmarshal(resp, &policies); err == nil {
				log.Printf("\nAPI Response received and parsed successfully")
				log.Printf("Response Kind: %s", policies.Kind)
				log.Printf("Generation: %d", policies.Generation)
				break
			}
			log.Printf("Error parsing WAF policies response: %v", err)
			lastErr = fmt.Errorf("JSON parsing error: %v", err)
			continue
		}

		lastErr = err
		errStr := err.Error()
		log.Printf("\nAPI request failed on attempt %d: %v", retry+1, err)

		// Determine if we should retry based on error type
		shouldRetry := false
		switch {
		case strings.Contains(strings.ToLower(errStr), "unauthorized"):
			log.Printf("Authentication Error: Please verify credentials and WAF module access permissions")
			// Don't retry auth errors
		case strings.Contains(strings.ToLower(errStr), "connection"):
			log.Printf("Connection Error: Unable to reach BIG-IP WAF endpoint")
			log.Printf("Please verify:\n1. Network connectivity\n2. BIG-IP management interface\n3. ASM module is provisioned and licensed")
		log.Printf("Attempting to verify ASM module status...")
		// Try to make a HEAD request to check if the endpoint exists
		headReq := &bigip.APIRequest{
			Method:      "HEAD",
			URL:         "mgmt/tm/asm/policies",
			ContentType: "application/json",
		}
		_, headErr := c.BigIP.APICall(headReq)
		if headErr != nil {
			log.Printf("ASM endpoint check failed: %v", headErr)
		} else {
			log.Printf("ASM endpoint exists but GET request failed - possible permission issue")
		}
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			log.Printf("Timeout Error: Request timed out")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "not found"):
			log.Printf("Endpoint Error: WAF/ASM endpoint not found")
			log.Printf("Please verify ASM module is provisioned on BIG-IP")
			// Don't retry 404 errors
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
			shouldRetry = true
		}

		if !shouldRetry || retry == maxRetries-1 {
			return nil, fmt.Errorf("failed to get WAF policies: %v", lastErr)
		}
	}

	var wafPolicies []*WAFPolicy
	log.Printf("\nProcessing %d WAF policies...", len(policies.Items))

	for _, policy := range policies.Items {
		log.Printf("\nProcessing policy:")
		log.Printf("  Name: %s", policy.Name)
		log.Printf("  ID: %s", policy.ID)
		log.Printf("  Type: %s", policy.Type)
		log.Printf("  Enforcement Mode: %s", policy.EnforcementMode)

		wafPolicy := &WAFPolicy{
			Name:             policy.Name,
			FullPath:         policy.FullPath,
			ID:               policy.ID,
			Description:      policy.Description,
			Active:           policy.Active,
			Type:             policy.Type,
			EnforcementMode:  policy.EnforcementMode,
			SignatureStaging: policy.SignatureStaging,
			VirtualServers:   policy.VirtualServers,
			BlockingMode:     policy.BlockingMode,
			PlaceSignatures:  policy.PlaceSignatures,
			SignatureSetings: policy.SignatureSetings,
			Kind:            policy.Kind,
			SelfLink:        policy.SelfLink,
		}
		wafPolicies = append(wafPolicies, wafPolicy)
	}

	log.Printf("\nFound and processed %d WAF policies successfully", len(wafPolicies))
	if len(wafPolicies) == 0 {
		log.Printf("\nWARNING: No WAF policies found. This could indicate that:")
		log.Printf("1. No WAF policies are configured")
		log.Printf("2. The ASM module might not be provisioned")
		log.Printf("3. The user might not have permissions to view WAF policies")
	} else {
		log.Printf("\nWAF Policies found:")
		for i, policy := range wafPolicies {
			log.Printf("[%d] %s (Type: %s, Mode: %s)", i+1, policy.Name, policy.Type, policy.EnforcementMode)
		}
	}
	return wafPolicies, nil
}

// GetWAFPolicyDetails retrieves detailed information about a specific WAF policy
func (c *Client) GetWAFPolicyDetails(policyName string) (*WAFPolicy, error) {
	if policyName == "" {
		return nil, fmt.Errorf("policy name cannot be empty")
	}
	log.Printf("\nAttempting to fetch details for WAF policy: %s", policyName)
	log.Printf("\n=== Starting GetWAFPolicyDetails Operation for policy: %s ===", policyName)
	log.Printf("Endpoint: /mgmt/tm/asm/policies")
	log.Printf("Method: GET")

	maxRetries := 3
	baseDelay := 5 * time.Second
	maxDelay := 30 * time.Second
	var lastErr error
	var policiesResp ASMPoliciesResponse

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			// Calculate exponential backoff delay
			backoffMultiplier := uint(1) << uint(retry-1)
			delay := baseDelay * time.Duration(backoffMultiplier)
			if delay > maxDelay {
				delay = maxDelay
			}
			log.Printf("Retry attempt %d/%d after %v delay (exponential backoff)...", retry+1, maxRetries, delay)
			time.Sleep(delay)
		}

		req := &bigip.APIRequest{
			Method:      "GET",
			URL:         fmt.Sprintf("mgmt/tm/asm/policies?$filter=name+eq+%s", policyName),
			ContentType: "application/json",
		}

		log.Printf("\nMaking API request to fetch details for WAF policy: %s", policyName)
		resp, err := c.BigIP.APICall(req)

		if err == nil {
			if err = json.Unmarshal(resp, &policiesResp); err == nil {
				log.Printf("\nAPI Response received and parsed successfully")
				break
			}
			log.Printf("Error parsing WAF policy details response: %v", err)
			lastErr = fmt.Errorf("JSON parsing error: %v", err)
			continue
		}

		lastErr = err
		errStr := err.Error()
		log.Printf("\nAPI request failed on attempt %d: %v", retry+1, err)

		// Determine if we should retry based on error type
		shouldRetry := false
		switch {
		case strings.Contains(strings.ToLower(errStr), "unauthorized"):
			log.Printf("Authentication Error: Please verify credentials and WAF module access permissions")
			// Don't retry auth errors
		case strings.Contains(strings.ToLower(errStr), "connection"):
			log.Printf("Connection Error: Unable to reach BIG-IP WAF endpoint")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			log.Printf("Timeout Error: Request timed out")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "not found"):
			log.Printf("Endpoint Error: WAF/ASM endpoint or policy not found")
			// Don't retry 404 errors
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
			shouldRetry = true
		}

		if !shouldRetry || retry == maxRetries-1 {
			return nil, fmt.Errorf("failed to get WAF policy details: %v", lastErr)
		}
	}

	if len(policiesResp.Items) == 0 {
		return nil, fmt.Errorf("WAF policy '%s' not found", policyName)
	}

	policy := policiesResp.Items[0]
	log.Printf("\nSuccessfully retrieved details for WAF policy: %s", policy.Name)
	log.Printf("Policy ID: %s", policy.ID)
	log.Printf("Type: %s", policy.Type)
	log.Printf("Status: %s", map[bool]string{true: "Active", false: "Inactive"}[policy.Active])

	return &WAFPolicy{
		Name:             policy.Name,
		FullPath:         policy.FullPath,
		ID:               policy.ID,
		Description:      policy.Description,
		Active:           policy.Active,
		Type:             policy.Type,
		EnforcementMode:  policy.EnforcementMode,
		SignatureStaging: policy.SignatureStaging,
		VirtualServers:   policy.VirtualServers,
		BlockingMode:     policy.BlockingMode,
		PlaceSignatures:  policy.PlaceSignatures,
		SignatureSetings: policy.SignatureSetings,
		Kind:            policy.Kind,
		SelfLink:        policy.SelfLink,
	}, nil
}

func (c *Client) GetVirtualServers() ([]VirtualServer, error) {
	log.Println("\n=== Starting GetVirtualServers Operation ===")
	log.Printf("Endpoint: /mgmt/tm/ltm/virtual")
	log.Printf("Method: GET")
	log.Printf("Authentication: Basic Auth (Username: %s)", c.Username)

	log.Println("\nMaking API request to fetch virtual servers...")
	vs, err := c.VirtualServers()
	if err != nil {
		log.Printf("\nERROR: Failed to fetch virtual servers")
		log.Printf("Error Type: %T", err)
		log.Printf("Error Message: %v", err)

		errStr := err.Error()
		switch {
		case strings.Contains(strings.ToLower(errStr), "unauthorized"):
			log.Printf("Authentication Error: Please verify credentials")
		case strings.Contains(strings.ToLower(errStr), "connection"):
			log.Printf("Connection Error: Unable to reach BIG-IP")
		case strings.Contains(strings.ToLower(errStr), "certificate"):
			log.Printf("TLS Certificate Error: Certificate validation failed")
		case strings.Contains(strings.ToLower(errStr), "no such host"):
			log.Printf("DNS Error: Unable to resolve BIG-IP hostname")
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			log.Printf("Timeout Error: Request took too long to complete")
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
		}
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	log.Println("\nAPI Response received successfully")

	var virtualServers []VirtualServer
	if vs != nil && vs.VirtualServers != nil {
		count := len(vs.VirtualServers)
		log.Printf("\nFound %d virtual server(s)", count)

		for i, v := range vs.VirtualServers {
			log.Printf("\nVirtual Server [%d/%d]:", i+1, count)
			log.Printf("  Name:        %s", v.Name)
			log.Printf("  Destination: %s", v.Destination)
			log.Printf("  Pool:        %s", v.Pool)
			log.Printf("  Status:      %s", map[bool]string{true: "Enabled", false: "Disabled"}[v.Enabled])
			vs := v // Create a copy to avoid referencing the loop variable
			virtualServers = append(virtualServers, VirtualServer{VirtualServer: &vs})
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

func (c *Client) GetPools() ([]Pool, map[string][]string, error) {
	pools, err := c.Pools()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pools: %v", err)
	}

	var poolList []Pool
	poolMembers := make(map[string][]string)

	for _, p := range pools.Pools {
		pool := p // Create a copy to avoid referencing the loop variable
		poolList = append(poolList, Pool{Pool: &pool})
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

func (c *Client) GetNodes() ([]Node, error) {
	nodes, err := c.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	var nodeList []Node
	for _, n := range nodes.Nodes {
		node := n // Create a copy to avoid referencing the loop variable
		nodeList = append(nodeList, Node{Node: &node})
	}
	return nodeList, nil
}