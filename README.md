# F5 BIG-IP Chat Interface

A Go-based chat interface for F5 BIG-IP management using OpenAI LLM. This tool provides a natural language interface to interact with F5 BIG-IP configurations through conversational commands.

## Features

- Natural language processing for BIG-IP management commands using OpenAI
- Interactive CLI interface for easy interaction
- Support for key BIG-IP components:
  - Virtual Servers (VIPs)
  - Server Pools
  - Backend Nodes
- Secure connection handling with TLS support
- Detailed logging for troubleshooting
- Human-friendly output formatting

## Prerequisites

1. Go 1.21 or later
2. OpenAI API key (for natural language processing)
3. Access to an F5 BIG-IP instance
4. Git (for cloning the repository)

## Quick Start

1. First, create a new repository on GitHub:
   - Go to github.com
   - Click "New Repository"
   - Name it "f5-bigip-chat"
   - Make it public
   - Don't initialize with any files

2. Clone and set up locally:
```bash
# Clone the repository
git clone https://github.com/[your-username]/f5-bigip-chat.git
cd f5-bigip-chat

# If you're setting up a new repository:
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/[your-username]/f5-bigip-chat.git
git push -u origin main
```

## Environment Setup

Set the following environment variables:

```bash
# BIG-IP Connection Details
export BIGIP_HOST="54.214.73.31:8443"     # Your BIG-IP host and port
export BIGIP_USERNAME="admin"              # Your BIG-IP username
export BIGIP_PASSWORD="F5testnet!"         # Your BIG-IP password

# OpenAI API Configuration
export OPENAI_API_KEY="your-openai-api-key"  # Your OpenAI API key
```

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

The application supports natural language queries. Here are some examples:

1. Virtual Server Management:
```
You: Show me all virtual servers
You: List the VIPs
You: What virtual servers are configured?
```

2. Pool Management:
```
You: List all pools
You: Show me the pool members
You: What's the status of our server pools?
```

3. Node Management:
```
You: Display all nodes
You: Show backend server status
You: List all backend nodes
```

## Project Structure

```
.
├── bigip/         # BIG-IP client implementation
├── chat/          # Chat interface logic
├── config/        # Configuration management
├── llm/           # LLM (OpenAI) integration
├── prompt/        # Prompt templates
├── utils/         # Utility functions
├── main.go        # Application entry point
└── README.md      # This file
```

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
