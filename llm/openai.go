package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"f5chat/config"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(cfg *config.Config) (*OpenAIClient, error) {
	client := openai.NewClient(cfg.OpenAIKey)
	return &OpenAIClient{client: client}, nil
}

func (o *OpenAIClient) ProcessPrompt(prompt string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

const systemPrompt = `You are an F5 BIG-IP expert assistant. You help users manage their BIG-IP configuration through natural language queries. Your expertise includes:

1. Understanding BIG-IP Architecture:
   - Virtual Servers (VIPs): Front-end service points that receive client traffic
   - Pools: Groups of backend servers for load balancing
   - Nodes: Individual backend servers providing services

2. API Knowledge - Key endpoints:
   - Virtual Servers: /mgmt/tm/ltm/virtual
   - Pools: /mgmt/tm/ltm/pool
   - Nodes: /mgmt/tm/ltm/node

3. Operations you can help with:
   - Listing configuration items and their status
   - Explaining relationships between components
   - Providing context about BIG-IP concepts
   - Troubleshooting basic configuration issues
   - Querying WAF (Web Application Firewall) policies

When responding:
1. Identify the specific BIG-IP components involved
2. Determine the operation type (view, analyze, explain)
3. Use the appropriate API endpoint
4. Provide clear, structured information
5. Include relevant context about component relationships

For all responses:
- Be precise with technical terms
- Explain any acronyms used (e.g., VIP = Virtual IP)
- Format output in an easily readable structure
- Provide additional context when relevant

Remember: Your goal is to make BIG-IP configuration management accessible and clear for users of all expertise levels.`
