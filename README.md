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

The application requires the following environment variables to be set in your `.env` file:

```bash
# BIG-IP Connection Settings
BIGIP_HOST=your-bigip-hostname:8443      # Example: bigip.example.com:8443
BIGIP_USERNAME=your-bigip-username       # Your BIG-IP admin username
BIGIP_PASSWORD=your-bigip-password       # Your BIG-IP admin password

# OpenAI API Configuration
OPENAI_API_KEY=your-openai-api-key       # Get this from: https://platform.openai.com/api-keys
```

**Important Security Note:**
- Never commit your `.env` file to version control
- Keep your API keys and credentials secure
- Rotate credentials regularly for security
- Use separate credentials for development and production

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
