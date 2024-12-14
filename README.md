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

1. Clone the repository:
```bash
git clone https://github.com/scshitole/chatf5.git
cd chatf5
```

2. Set up environment variables in your terminal:
```bash
# BIG-IP Connection Details
export BIGIP_HOST='your-bigip-host:8443'
export BIGIP_USERNAME='your-bigip-username'
export BIGIP_PASSWORD='your-bigip-password'

# OpenAI API Configuration
export OPENAI_API_KEY='your-openai-api-key'
```

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

The application requires the following environment variables:

- `BIGIP_HOST`: The hostname/IP and port of your F5 BIG-IP instance
- `BIGIP_USERNAME`: Your BIG-IP username
- `BIGIP_PASSWORD`: Your BIG-IP password
- `OPENAI_API_KEY`: Your OpenAI API key for natural language processing

You can set these variables as shown in the Quick Start section above.

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
