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

func FormatWAFPolicies(policies []*bigip.WAFPolicy) string {
	var sb strings.Builder
	sb.WriteString("\n=== WAF (Web Application Firewall) Policies ===\n")
	
	if len(policies) == 0 {
		sb.WriteString("\nNo WAF policies are currently configured on this BIG-IP system.\n")
		sb.WriteString("Note: WAF policies protect web applications from various attacks like SQL injection, XSS, etc.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("\nFound %d WAF Policies:\n", len(policies)))
	
	for i, policy := range policies {
		sb.WriteString(fmt.Sprintf("\n[%d] WAF Policy Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
		sb.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "Active", false: "Inactive"}[policy.Active]))
		if policy.EnforcementMode != "" {
			sb.WriteString(fmt.Sprintf("Enforcement Mode: %s\n", policy.EnforcementMode))
		}
		if policy.Type != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", policy.Type))
		}
		if len(policy.VirtualServers) > 0 {
			sb.WriteString("Associated Virtual Servers:\n")
			for _, vs := range policy.VirtualServers {
				sb.WriteString(fmt.Sprintf("- %s\n", vs))
			}
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
	sb.WriteString(fmt.Sprintf("\n=== WAF Policy Details: %s ===\n", policy.Name))
	sb.WriteString("----------------------------------------\n")
	
	sb.WriteString(fmt.Sprintf("Name: %s\n", policy.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", policy.ID))
	sb.WriteString(fmt.Sprintf("Type: %s\n", policy.Type))
	sb.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "Active", false: "Inactive"}[policy.Active]))
	sb.WriteString(fmt.Sprintf("Enforcement Mode: %s\n", policy.EnforcementMode))
	
	if policy.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", policy.Description))
	}
	
	if policy.SignatureStaging {
		sb.WriteString("Signature Mode: Staging\n")
	} else {
		sb.WriteString("Signature Mode: Production\n")
	}
	
	if len(policy.VirtualServers) > 0 {
		sb.WriteString("\nAssociated Virtual Servers:\n")
		for _, vs := range policy.VirtualServers {
			sb.WriteString(fmt.Sprintf("- %s\n", vs))
		}
	}
	
	sb.WriteString("\nConfiguration Path: " + policy.FullPath + "\n")
	
	return sb.String()
}
