package utils

import (
	"fmt"
	"strings"

	"github.com/f5devcentral/go-bigip"
)

func FormatVirtualServers(vs []bigip.VirtualServer) string {
	var sb strings.Builder
	sb.WriteString("\n=== Virtual Servers (VIPs) ===\n")
	
	if len(vs) == 0 {
		sb.WriteString("No virtual servers configured.\n")
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

func FormatPools(pools []bigip.Pool, poolMembers map[string][]string) string {
	var sb strings.Builder
	sb.WriteString("\n=== Server Pools ===\n")

	if len(pools) == 0 {
		sb.WriteString("No pools configured.\n")
		return sb.String()
	}

	for i, p := range pools {
		sb.WriteString(fmt.Sprintf("\n[%d] Pool Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name:         %s\n", p.Name))
		sb.WriteString(fmt.Sprintf("Load Balance: %s\n", p.LoadBalancingMode))
		sb.WriteString(fmt.Sprintf("Monitor:      %s\n", p.Monitor))
		
		// Display pool members
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

func FormatNodes(nodes []bigip.Node) string {
	var sb strings.Builder
	sb.WriteString("\n=== Backend Nodes ===\n")

	if len(nodes) == 0 {
		sb.WriteString("No nodes configured.\n")
		return sb.String()
	}

	for i, n := range nodes {
		sb.WriteString(fmt.Sprintf("\n[%d] Node Details:\n", i+1))
		sb.WriteString("----------------------------------------\n")
		sb.WriteString(fmt.Sprintf("Name:        %s\n", n.Name))
		sb.WriteString(fmt.Sprintf("Address:     %s\n", n.Address))
		sb.WriteString(fmt.Sprintf("State:       %s\n", n.State))
		
		// Add connection limits if configured
		if n.ConnectionLimit > 0 {
			sb.WriteString(fmt.Sprintf("Conn Limit:  %d\n", n.ConnectionLimit))
		}
		
		// Add dynamic ratio if configured
		if n.DynamicRatio > 0 {
			sb.WriteString(fmt.Sprintf("Dyn Ratio:   %d\n", n.DynamicRatio))
		}
		
		if n.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", n.Description))
		}
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String()
}
