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
   - Policy listing and details
   - Virtual server associations
   - Security settings viewing
   - Enforcement mode status

4. Security Features:
   - TLS 1.2+ support
   - Certificate validation
   - Secure credential handling
   - Detailed audit logging

5. Technical Capabilities:
   - REST API integration
   - Automatic retry mechanisms
   - Error handling and recovery
   - Comprehensive logging system

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

2. Create a `.env` file in the project root:
```bash
# Copy the .env.example file
cp .env.example .env

# Edit .env file with your credentials
# Replace the placeholder values with your actual credentials
# DO NOT commit this file to version control
```

The `.env` file should contain your actual credentials following the format in the Environment Variables section above.

3. Install Go (if not already installed):
```bash
brew install go  # On macOS using Homebrew
```

4. Install dependencies and run:
```bash
go mod download
go mod verify
go run main.go
```

## Environment Variables

The application requires the following environment variables to be set in your `.env` file. Reference: iControl REST API User Guide 14.1.0, Chapter 2: REST API Authentication.

```bash
# BIG-IP Connection Settings
BIGIP_HOST=<hostname>:<port>       # Example: bigip.example.com:8443
                                  # Port 8443 is the default HTTPS management port
BIGIP_USERNAME=<username>         # BIG-IP admin or user with appropriate role
                                  # Requires minimum "Resource Admin" role
BIGIP_PASSWORD=<password>         # BIG-IP user's password
                                  # Use strong passwords with special characters

# OpenAI API Configuration
OPENAI_API_KEY=<api-key>         # OpenAI API key for NLP processing
                                  # Get this from: https://platform.openai.com/api-keys
```

**Important Security Notes:**
1. Environment File Security:
   - Never commit `.env` file to version control
   - Add `.env` to your `.gitignore` file
   - Keep a sanitized `.env.example` for reference

2. BIG-IP Access Security:
   - Use HTTPS (port 8443) for management access
   - Create dedicated API user accounts
   - Apply proper role-based access control (RBAC)
   - Refer to F5 documentation for security best practices

3. Credential Management:
   - Rotate credentials regularly
   - Use strong passwords with special characters
   - Keep API keys and credentials secure
   - Use separate credentials for development and production
   - Monitor and audit API access regularly

## Installation

1. Install dependencies:
```bash
go mod download
```

2. Verify the installation:
```bash
go mod verify
```

## Running the Application

Start the application:
```bash
go run main.go
```

## Usage Examples

The application supports natural language queries based on iControl REST API endpoints. Here are some examples:

1. Virtual Server Management (Endpoint: /mgmt/tm/ltm/virtual):
```
You: Show me all virtual servers
You: List the VIPs and their status
You: What virtual servers are currently enabled?
You: Display VIPs with their pools
```

2. Pool Management (Endpoint: /mgmt/tm/ltm/pool):
```
You: List all pools and their members
You: Show me the pool health status
You: What's the load balancing method for each pool?
You: Display pools with their monitoring settings
```

3. Node Management (Endpoint: /mgmt/tm/ltm/node):
```
You: Display all backend nodes
You: Show node health status
You: List all nodes and their addresses
You: What's the current state of our nodes?
```

4. WAF Policy Management (Endpoint: /mgmt/tm/asm/policies):
```
You: List all WAF policies
You: Show policy details for demo
You: Display WAF policies and their virtual servers
You: What's the enforcement mode of our WAF policies?
```

Each command interfaces with specific BIG-IP REST endpoints and returns formatted, human-readable responses. For more details on the API endpoints, refer to the iControl REST API documentation.

## Project Structure

```
.
├── bigip/         # BIG-IP client implementation
│   └── client.go  # REST API client with retry logic
├── chat/          # Chat interface logic
│   └── interface.go # Natural language command processing
├── config/        # Configuration management
│   └── config.go  # Environment and settings handler
├── llm/           # LLM (OpenAI) integration
│   └── openai.go  # OpenAI API client
├── prompt/        # Prompt templates
│   └── templates.go # System prompts and templates
├── utils/         # Utility functions
│   └── formatter.go # Output formatting
├── main.go        # Application entry point
├── setup.sh       # Environment setup script
├── .env.example   # Environment template
└── README.md      # Documentation
```

Each component is designed according to the iControl REST API architecture, ensuring proper separation of concerns and maintainable code structure.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

MIT License

## Support

For support, please open an issue in the GitHub repository.
