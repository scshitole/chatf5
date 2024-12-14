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
type SignatureStatus struct {
	ID              string `json:"id"`
	SignatureID     string `json:"signatureId"`
	SignatureName   string `json:"signatureName"`
	Enabled         bool   `json:"enabled"`
	PerformStaging  bool   `json:"performStaging"`
	Block           bool   `json:"block"`
	Description     string `json:"description,omitempty"`
	SignatureType   string `json:"signatureType,omitempty"`
	AccuracyLevel   string `json:"accuracy,omitempty"`
	RiskLevel       string `json:"riskLevel,omitempty"`
	PolicyName      string `json:"policyName,omitempty"`
	Context         string `json:"context,omitempty"`
}

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
	// Parse host and port according to iControl REST API specs
	hostParts := strings.Split(strings.TrimSpace(cfg.BigIPHost), ":")
	host := hostParts[0]
	port := "443" // default HTTPS port per F5 documentation
	if len(hostParts) > 1 {
		port = hostParts[1]
	}

	// Construct proper URL with required format: https://<hostname>/mgmt/tm/<module>
	baseURL := fmt.Sprintf("https://%s:%s", host, port)

	// Create configuration for BIG-IP session
	config := &bigip.Config{
		Address:  baseURL,
		Username: cfg.BigIPUsername,
		Password: cfg.BigIPPassword,
	}

	bigipClient := bigip.NewSession(config)

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
	log.Printf("Testing connection to BIG-IP...")

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
			vs, testErr := bigipClient.VirtualServers()
			if testErr == nil && vs != nil {
				log.Printf("Connected to BIG-IP successfully")
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
// Reference: iControl REST API User Guide 14.1.0
// Endpoint: /mgmt/tm/asm/policies
// Method: GET
// Required Role: Administrator or Resource Administrator
// Response Format: Collection of ASM policy objects
func (c *Client) GetWAFPolicies() ([]*WAFPolicy, error) {
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

		// Error handling based on F5 iControl REST API error codes
		// Reference: iControl REST API User Guide 14.1.0, Chapter 4: Response Codes
		shouldRetry := false
		switch {
		case strings.Contains(strings.ToLower(errStr), "401"):
			// 401: Authentication failed
			log.Printf("Authentication Error (401): Invalid credentials or token expired")
			log.Printf("Action: Verify username/password or renew authentication token")
		case strings.Contains(strings.ToLower(errStr), "403"):
			// 403: Authorization failed
			log.Printf("Authorization Error (403): Insufficient permissions for the operation")
			log.Printf("Action: Verify user role assignments and partition access")
		case strings.Contains(strings.ToLower(errStr), "404"):
			// 404: Resource not found
			log.Printf("Resource Not Found (404): The requested resource does not exist")
			log.Printf("Action: Verify the resource path and partition accessibility")
			shouldRetry = false
		case strings.Contains(strings.ToLower(errStr), "409"):
			// 409: Conflict in resource state
			log.Printf("Resource Conflict (409): The request conflicts with the current state")
			log.Printf("Action: Verify resource state and try again")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "connection"):
			// Connection-level errors
			log.Printf("Connection Error: Unable to reach BIG-IP")
			log.Printf("Action: Verify network connectivity and BIG-IP availability")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "certificate"):
			// SSL/TLS errors
			log.Printf("TLS Certificate Error: Certificate validation failed")
			log.Printf("Action: Verify SSL certificate or use SSL verification bypass for testing")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			// Timeout errors
			log.Printf("Timeout Error (408): Request exceeded time limit")
			log.Printf("Action: Verify BIG-IP load and network latency")
			shouldRetry = true
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
			log.Printf("Action: Check BIG-IP logs for detailed information")
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

// GetWAFPolicyDetails retrieves details for a specific WAF policy
// Reference: iControl REST API Guide 14.1.0, Chapter 7: Application Security Management
// Endpoint: /mgmt/tm/asm/policies
// Method: GET with policy name filter
func (c *Client) GetWAFPolicyDetails(policyName string) (*WAFPolicy, error) {
	if policyName == "" {
		return nil, fmt.Errorf("policy name cannot be empty")
	}

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

		// Error handling based on F5 iControl REST API error codes
		// Reference: iControl REST API User Guide 14.1.0, Chapter 4: Response Codes
		shouldRetry := false
		switch {
		case strings.Contains(strings.ToLower(errStr), "401"):
			// 401: Authentication failed
			log.Printf("Authentication Error (401): Invalid credentials or token expired")
			log.Printf("Action: Verify username/password or renew authentication token")
		case strings.Contains(strings.ToLower(errStr), "403"):
			// 403: Authorization failed
			log.Printf("Authorization Error (403): Insufficient permissions for the operation")
			log.Printf("Action: Verify user role assignments and partition access")
		case strings.Contains(strings.ToLower(errStr), "404"):
			// 404: Resource not found
			log.Printf("Resource Not Found (404): The requested resource does not exist")
			log.Printf("Action: Verify the resource path and partition accessibility")
			shouldRetry = false
		case strings.Contains(strings.ToLower(errStr), "409"):
			// 409: Conflict in resource state
			log.Printf("Resource Conflict (409): The request conflicts with the current state")
			log.Printf("Action: Verify resource state and try again")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "connection"):
			// Connection-level errors
			log.Printf("Connection Error: Unable to reach BIG-IP")
			log.Printf("Action: Verify network connectivity and BIG-IP availability")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "certificate"):
			// SSL/TLS errors
			log.Printf("TLS Certificate Error: Certificate validation failed")
			log.Printf("Action: Verify SSL certificate or use SSL verification bypass for testing")
			shouldRetry = true
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			// Timeout errors
			log.Printf("Timeout Error (408): Request exceeded time limit")
			log.Printf("Action: Verify BIG-IP load and network latency")
			shouldRetry = true
		default:
			log.Printf("Unhandled error type - Full error: %v", err)
			log.Printf("Action: Check BIG-IP logs for detailed information")
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
	vs, err := c.VirtualServers()
	if err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(strings.ToLower(errStr), "unauthorized"):
			return nil, fmt.Errorf("authentication failed: please verify credentials")
		case strings.Contains(strings.ToLower(errStr), "connection"):
			return nil, fmt.Errorf("connection error: unable to reach BIG-IP")
		case strings.Contains(strings.ToLower(errStr), "certificate"):
			return nil, fmt.Errorf("TLS certificate error: certificate validation failed")
		case strings.Contains(strings.ToLower(errStr), "no such host"):
			return nil, fmt.Errorf("DNS error: unable to resolve BIG-IP hostname")
		case strings.Contains(strings.ToLower(errStr), "timeout"):
			return nil, fmt.Errorf("timeout error: request took too long to complete")
		default:
			return nil, fmt.Errorf("API request failed: %v", err)
		}
	}

	var virtualServers []VirtualServer
	if vs != nil && vs.VirtualServers != nil {
		for _, v := range vs.VirtualServers {
			vs := v // Create a copy to avoid referencing the loop variable
			virtualServers = append(virtualServers, VirtualServer{VirtualServer: &vs})
		}
	}

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


// GetPolicySignatureStatus retrieves signature status information for a specific WAF policy
// Reference: iControl REST API Guide 14.1.0, Chapter 7: Application Security Management
// Endpoint: /mgmt/tm/asm/signature-statuses
func (c *Client) GetPolicySignatureStatus(policyID string) ([]SignatureStatus, error) {
	if policyID == "" {
		return nil, fmt.Errorf("policy ID cannot be empty")
	}

	type SignatureResponse struct {
		Items []SignatureStatus `json:"items"`
	}

	var signatures SignatureResponse
	req := &bigip.APIRequest{
		Method:      "GET",
		URL:         "mgmt/tm/asm/signature-statuses",
		ContentType: "application/json",
	}

	log.Printf("\nMaking API request to fetch signature status...")
	resp, err := c.BigIP.APICall(req)
	if err != nil {
		log.Printf("Error fetching signature status: %v", err)
		return nil, fmt.Errorf("failed to get signature status: %v", err)
	}

	if err = json.Unmarshal(resp, &signatures); err != nil {
		log.Printf("Error parsing signature status response: %v", err)
		return nil, fmt.Errorf("failed to parse signature status: %v", err)
	}

	log.Printf("\nFound %d signatures for policy", len(signatures.Items))
	for i, sig := range signatures.Items {
		log.Printf("Signature [%d]:", i+1)
		log.Printf("  ID: %s", sig.SignatureID)
		log.Printf("  Name: %s", sig.SignatureName)
		log.Printf("  Enabled: %v", sig.Enabled)
		log.Printf("  Staging: %v", sig.PerformStaging)
	}

	return signatures.Items, nil
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