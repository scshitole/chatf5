package utils

import (
	"fmt"
	"strings"

	"f5chat/bigip"
)

// Type aliases for bigip package types
type (
	VirtualServer = bigip.VirtualServer
	Pool         = bigip.Pool
	Node         = bigip.Node
	WAFPolicy    = bigip.WAFPolicy
)

func FormatVirtualServers(vs []VirtualServer) string {
	var sb strings.Builder
	sb.WriteString("\n=== Virtual Servers (VIPs) ===\n")
	
	if len(vs) == 0 {
		sb.WriteString("\nNo virtual servers are currently configured.\n")
		return sb.String()
	}

	for i, v := range vs {
		sb.WriteString(fmt.Sprintf("\n[%d] Virtual Server Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name:        %s\n", v.Name))
		sb.WriteString(fmt.Sprintf("Destination: %s\n", v.Destination))
		sb.WriteString(fmt.Sprintf("Pool:        %s\n", v.Pool))
		status := "Enabled"
		if !v.Enabled {
			status = "Disabled"
		}
		sb.WriteString(fmt.Sprintf("Status:      %s\n", status))
		if v.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", v.Description))
		}
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String()
}

func FormatPools(pools []Pool, poolMembers map[string][]string) string {
	var sb strings.Builder
	sb.WriteString("\n=== Server Pools ===\n")

	if len(pools) == 0 {
		sb.WriteString("\nNo server pools are currently configured.\n")
		return sb.String()
	}

	for i, p := range pools {
		sb.WriteString(fmt.Sprintf("\n[%d] Pool Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name:         %s\n", p.Name))
		sb.WriteString(fmt.Sprintf("Load Balance: %s\n", p.LoadBalancingMode))
		sb.WriteString(fmt.Sprintf("Monitor:      %s\n", p.Monitor))
		
		sb.WriteString("\nPool Members:\n")
		if members, ok := poolMembers[p.Name]; ok && len(members) > 0 {
			for j, m := range members {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", j+1, m))
			}
		} else {
			sb.WriteString("  No members configured\n")
		}
		
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("\nDescription: %s\n", p.Description))
		}
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String()
}

func FormatNodes(nodes []Node) string {
	var sb strings.Builder
	sb.WriteString("\n=== Backend Nodes ===\n")
	
	if len(nodes) == 0 {
		sb.WriteString("\nNo backend nodes are currently configured.\n")
		return sb.String()
	}

	for i, node := range nodes {
		sb.WriteString(fmt.Sprintf("\n[%d] Node Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name:    %s\n", node.Name))
		sb.WriteString(fmt.Sprintf("Address: %s\n", node.Address))
		sb.WriteString(fmt.Sprintf("State:   %s\n", node.State))
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String()
}

// FormatWAFPolicies formats WAF/ASM policies according to iControl REST API structure
// Reference: iControl REST API v14.1.0, Chapter 7: Application Security Management
func FormatWAFPolicies(policies []*WAFPolicy) string {
	var sb strings.Builder
	sb.WriteString("\n=== WAF (Web Application Firewall) Policies ===\n")
	sb.WriteString("Reference: iControl REST API v14.1.0\n")
	
	if len(policies) == 0 {
		sb.WriteString("\nNo WAF policies are currently configured on this BIG-IP system.\n")
		sb.WriteString("\nNote: WAF policies protect web applications from:")
		sb.WriteString("\n- SQL injection attacks")
		sb.WriteString("\n- Cross-site scripting (XSS)")
		sb.WriteString("\n- Request/Response validation")
		sb.WriteString("\n- Protocol compliance")
		sb.WriteString("\n- Other OWASP Top 10 vulnerabilities\n")
		sb.WriteString("\nTo configure a WAF policy, use the BIG-IP Configuration utility or API.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("\nFound %d WAF Policies:\n", len(policies)))
	
	for i, policy := range policies {
		sb.WriteString(fmt.Sprintf("\n[%d] WAF Policy Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
		sb.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "Active", false: "Inactive"}[policy.Active]))
		
		// Display Virtual Server associations prominently
		if len(policy.VirtualServers) > 0 {
			sb.WriteString("\nApplied to Virtual Servers:\n")
			for _, vs := range policy.VirtualServers {
				sb.WriteString(fmt.Sprintf("- %s\n", vs))
			}
			sb.WriteString("\n")
		} else {
			sb.WriteString("\nNot currently applied to any Virtual Servers\n\n")
		}
		
		if policy.EnforcementMode != "" {
			sb.WriteString(fmt.Sprintf("Enforcement Mode: %s\n", policy.EnforcementMode))
			if policy.EnforcementMode == "blocking" {
				sb.WriteString("  (Actively blocking detected violations)\n")
			} else if policy.EnforcementMode == "transparent" {
				sb.WriteString("  (Monitoring mode - logging only)\n")
			}
		}
		
		if policy.Type != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", policy.Type))
		}
		
		sb.WriteString(fmt.Sprintf("Signature Staging: %v\n", map[bool]string{
			true:  "Enabled (New signatures in staging mode)",
			false: "Disabled (All signatures in production)",
		}[policy.SignatureStaging]))
		
		if len(policy.VirtualServers) > 0 {
			sb.WriteString("\nAssociated Virtual Servers:\n")
			for _, vs := range policy.VirtualServers {
				sb.WriteString(fmt.Sprintf("- %s\n", vs))
			}
		}
		
		if policy.Description != "" {
			sb.WriteString(fmt.Sprintf("\nDescription: %s\n", policy.Description))
		}
		
		sb.WriteString("----------------------------------------\n")
	}

	sb.WriteString("\nNote: WAF policies are configured to protect web applications ")
	sb.WriteString("from various attacks such as SQL injection, cross-site scripting (XSS), ")
	sb.WriteString("and other OWASP Top 10 vulnerabilities.\n")
	sb.WriteString("\nTip: To see detailed information about a specific policy, ")
	sb.WriteString("ask about 'policy details for [policy name]'\n")

	return sb.String()
}

func FormatWAFPolicyDetails(policy *bigip.WAFPolicy) string {
	var sb strings.Builder
	sb.WriteString("\n=== WAF (Web Application Firewall) Policy Details ===\n")
	sb.WriteString("Reference: iControl REST API v14.1.0, Chapter 7: Application Security Management\n")
	sb.WriteString("----------------------------------------\n")
	
	// Basic Information with Improved Formatting
	sb.WriteString("BASIC INFORMATION:\n")
	sb.WriteString(fmt.Sprintf("• Name: %s\n", policy.Name))
	sb.WriteString(fmt.Sprintf("• Full Path: %s\n", policy.FullPath))
	sb.WriteString(fmt.Sprintf("• ID: %s\n", policy.ID))
	
	// Policy Status and Type with Enhanced Details
	sb.WriteString("\nSTATUS AND CONFIGURATION:\n")
	sb.WriteString(fmt.Sprintf("• Active: %s\n", map[bool]string{true: "Yes (Policy is enforcing security rules)", 
		false: "No (Policy is inactive)"}[policy.Active]))
	if policy.Type != "" {
		sb.WriteString(fmt.Sprintf("• Type: %s\n", policy.Type))
	}
	
	// Enforcement Configuration
	sb.WriteString("\nENFORCEMENT SETTINGS:\n")
	if policy.EnforcementMode != "" {
		sb.WriteString(fmt.Sprintf("• Mode: %s\n", policy.EnforcementMode))
		switch policy.EnforcementMode {
		case "blocking":
			sb.WriteString("  ↳ Policy is actively preventing detected violations\n")
			sb.WriteString("  ↳ Malicious requests are blocked in real-time\n")
		case "transparent":
			sb.WriteString("  ↳ Policy is in monitoring/learning mode\n")
			sb.WriteString("  ↳ Violations are logged but not blocked\n")
		}
	}
	
	// Signature Configuration with Detailed Explanation
	sb.WriteString("\nSIGNATURE SETTINGS:\n")
	sb.WriteString(fmt.Sprintf("• Staging: %s\n", map[bool]string{
		true:  "Enabled - New signatures are in staging mode",
		false: "Disabled - All signatures are in production",
	}[policy.SignatureStaging]))
	if policy.SignatureStaging {
		sb.WriteString("  ↳ New attack signatures are monitored without blocking\n")
		sb.WriteString("  ↳ Helps prevent false positives with new signatures\n")
	}
	if policy.BlockingMode != "" {
		sb.WriteString(fmt.Sprintf("• Blocking Mode: %s\n", policy.BlockingMode))
	}
	
	// Virtual Server Associations with Status
	sb.WriteString("\nVIRTUAL SERVER ASSOCIATIONS:\n")
	if len(policy.VirtualServers) > 0 {
		for _, vs := range policy.VirtualServers {
			sb.WriteString(fmt.Sprintf("• %s\n", vs))
		}
		sb.WriteString("\nNote: This policy is actively protecting the above virtual servers\n")
	} else {
		sb.WriteString("• Not currently applied to any Virtual Servers\n")
		sb.WriteString("Note: Policy is configured but not actively protecting any services\n")
	}
	
	// Additional Information
	if policy.Description != "" {
		sb.WriteString("\nDESCRIPTION:\n")
		sb.WriteString(fmt.Sprintf("%s\n", policy.Description))
	}
	
	// API Information
	sb.WriteString("\nAPI REFERENCE:\n")
	sb.WriteString(fmt.Sprintf("• Self Link: %s\n", policy.SelfLink))
	sb.WriteString(fmt.Sprintf("• Kind: %s\n", policy.Kind))
	sb.WriteString(fmt.Sprintf("• Policy ID: %s\n", policy.ID))
	
	sb.WriteString("\nTIP: Use this policy ID for direct API requests and automation\n")
	sb.WriteString("----------------------------------------\n")
	
	return sb.String()
}

// FormatSignatureStatus formats the signature status information for display
func FormatSignatureStatus(signatures []bigip.SignatureStatus) string {
	var sb strings.Builder
	sb.WriteString("\n=== WAF Policy Signature Status ===\n")
	sb.WriteString("Reference: iControl REST API v14.1.0, Chapter 7: ASM Signatures\n")
	sb.WriteString("----------------------------------------\n")

	if len(signatures) == 0 {
		sb.WriteString("\nNo signatures found for this policy.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("\nFound %d Signatures:\n", len(signatures)))

	for i, sig := range signatures {
		sb.WriteString(fmt.Sprintf("\n[%d] Signature Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", sig.Name))
		sb.WriteString(fmt.Sprintf("Signature ID: %s\n", sig.SignatureID))
		sb.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{
			true:  "Enabled",
			false: "Disabled",
		}[sig.Enabled]))
		sb.WriteString(fmt.Sprintf("Staging: %s\n", map[bool]string{
			true:  "Yes (Learning Mode)",
			false: "No (Enforcement Mode)",
		}[sig.Staging]))
		sb.WriteString(fmt.Sprintf("Blocking: %s\n", map[bool]string{
			true:  "Enabled (Violations are blocked)",
			false: "Disabled (Monitoring only)",
		}[sig.BlockingEnabled]))

		if sig.SignatureType != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", sig.SignatureType))
		}
		if sig.AccuracyLevel != "" {
			sb.WriteString(fmt.Sprintf("Accuracy: %s\n", sig.AccuracyLevel))
		}
		if sig.RiskLevel != "" {
			sb.WriteString(fmt.Sprintf("Risk Level: %s\n", sig.RiskLevel))
		}
		if sig.Description != "" {
			sb.WriteString(fmt.Sprintf("\nDescription: %s\n", sig.Description))
		}
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String()
}