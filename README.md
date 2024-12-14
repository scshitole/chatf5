# F5 BIG-IP Chat Interface

A Go-based chat interface for F5 BIG-IP management using OpenAI LLM. This tool provides a natural language interface to interact with F5 BIG-IP configurations through conversational commands.

## Features

Reference: iControl REST API User Guide 14.1.0

1. Natural Language Interface:
   - OpenAI-powered command processing
   - Conversational BIG-IP management
   - Context-aware responses
   - Human-friendly output formatting

2. BIG-IP Component Management:
   - Virtual Servers (VIPs)
     * Configuration viewing
     * Status monitoring
     * Load balancing settings
   - Server Pools
     * Member management
     * Health monitoring
     * Load balancing methods
   - Backend Nodes
     * Status tracking
     * Address management
     * Health state monitoring

3. WAF Policy Management:
   - Policy listing and details (endpoint: /mgmt/tm/asm/policies)
   - Virtual server associations and status
   - Security settings and enforcement modes
   - Policy state monitoring (active/inactive)
   - Detailed policy information retrieval

4. Security Features:
   - TLS 1.2+ support
   - Certificate validation
   - Secure credential handling
   - Detailed audit logging

## Prerequisites

1. Go 1.21 or later
2. OpenAI API key (for natural language processing)
3. Access to an F5 BIG-IP instance
4. Git (for cloning the repository)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/scshitole/chatf5.git
cd chatf5
```

2. Create and configure environment:
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your credentials
# DO NOT commit the .env file
```

3. Install dependencies and run:
```bash
go mod download
go run main.go
```

## Environment Variables

Configure these in your `.env` file (see `.env.example` for template):

```bash
# BIG-IP Connection Settings
BIGIP_HOST=bigip.example.com:8443    # Your BIG-IP hostname and management port
                                     # Default management port is 8443 for HTTPS
BIGIP_USERNAME=admin                 # BIG-IP username with appropriate permissions
                                     # Requires minimum "Resource Administrator" role
BIGIP_PASSWORD=your-secure-password  # BIG-IP password (use strong password)
                                     # Minimum 8 characters with mixed case and symbols

# OpenAI API Configuration
OPENAI_API_KEY=your-api-key-here     # Your OpenAI API key
                                     # Get from: https://platform.openai.com/
```

**Security Requirements:**
1. Environment File Security:
   - Never commit `.env` to version control
   - Keep `.env.example` as a template only
   - Secure storage of actual credentials
   - Regular credential rotation

2. Access Control:
   - Use HTTPS (port 8443) for management
   - Implement proper RBAC on BIG-IP
   - Monitor API access logs
   - Use strong passwords with special characters

3. API Security:
   - Protect API keys and credentials
   - Regular key rotation
   - Monitor API usage
   - Keep OpenAI API key secure

## Usage Examples

The interface supports natural language queries. Examples:

1. Virtual Server Management:
```
You: Show all virtual servers
You: List VIPs and their status
You: Display enabled virtual servers
```

2. Pool Management:
```
You: List all pools and members
You: Show pool health status
You: Display pool monitoring settings
```

3. Node Management:
```
You: Show all backend nodes
You: Display node health status
You: List node addresses
```

4. WAF Policy Management (Endpoint: /mgmt/tm/asm/policies):
```
You: List WAF policies                    # Lists all WAF policies
You: Show policy details for demo         # Shows detailed configuration
You: Display WAF virtual server mappings  # Shows policy-VS associations
You: What's the enforcement mode?         # Shows policy enforcement status
```

Reference: iControl REST API Guide 14.1.0
- Endpoint: /mgmt/tm/asm/policies
- Authentication: Basic Auth
- Protocol: HTTPS required
- Content-Type: application/json

## Project Structure

```
.
├── bigip/         # BIG-IP client implementation
├── chat/          # Chat interface logic
├── config/        # Configuration management
├── llm/           # OpenAI integration
├── prompt/        # System prompts
├── utils/         # Utility functions
├── main.go        # Entry point
└── README.md      # Documentation
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/NewFeature`)
3. Commit changes (`git commit -m 'Add NewFeature'`)
4. Push to branch (`git push origin feature/NewFeature`)
5. Open Pull Request

## License

MIT License

## Support

For support, please open an issue in the GitHub repository.
