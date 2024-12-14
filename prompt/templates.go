package prompt

const (
	WAFPolicyListTemplate = `To list WAF policies, I'll need to:
1. Query the /mgmt/tm/asm/policies endpoint
2. Format and display the results including:
   - Name: The unique identifier of the WAF policy
   - Status: Current operational status
   - Type: Type of WAF policy
Additional Information:
- WAF policies protect web applications from attacks
- They contain rules and settings for web application security`

	VirtualServerListTemplate = `To list virtual servers (VIPs), I'll need to:
1. Query the /mgmt/tm/ltm/virtual endpoint
2. Format and display the results including:
   - Name: The unique identifier of the virtual server
   - Destination: IP:Port combination where the virtual server listens
   - Pool: Associated server pool name
   - Status: Current operational status (enabled/disabled)
Additional Information:
- Virtual servers are the primary ingress points for client traffic
- They distribute incoming connections across backend pool members`

	PoolListTemplate = `To list server pools, I'll need to:
1. Query the /mgmt/tm/ltm/pool endpoint
2. Format and display the results including:
   - Name: The unique identifier of the pool
   - Members: List of backend servers (nodes) in the pool
   - Monitor: Health check configuration
   - Status: Aggregate status of pool members
Additional Information:
- Pools manage groups of backend servers
- They handle load balancing and health monitoring of members`

	NodeListTemplate = `To list backend nodes, I'll need to:
1. Query the /mgmt/tm/ltm/node endpoint
2. Format and display the results including:
   - Name: The unique identifier of the node
   - Address: IP address of the backend server
   - Status: Current operational status
   - Monitor Status: Health check status
Additional Information:
- Nodes represent actual backend servers
- They can be members of multiple pools
- Monitor status indicates their availability`
)

func GetPromptTemplate(operation string) string {
	templates := map[string]string{
		"virtual_servers": VirtualServerListTemplate,
		"pools":          PoolListTemplate,
		"nodes":          NodeListTemplate,
		"waf_policies":   WAFPolicyListTemplate,
	}

	if template, exists := templates[operation]; exists {
		return template
	}
	return ""
}
