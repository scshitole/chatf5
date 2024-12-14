#!/bin/bash

# Setup script for F5 BIG-IP Chat Interface

# Check if .env file exists
if [ -f .env ]; then
    echo "Warning: .env file already exists. Remove it first if you want to create a new one."
    exit 1
fi

# Copy example environment file
echo "Creating .env file from template..."
cp .env.example .env

echo "Environment file created successfully!"
echo "Please edit .env file with your actual credentials:"
echo "1. Update BIGIP_HOST with your F5 BIG-IP hostname and port"
echo "2. Set BIGIP_USERNAME and BIGIP_PASSWORD"
echo "3. Add your OpenAI API key (OPENAI_API_KEY)"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your credentials"
echo "2. Run 'go mod download' to install dependencies"
echo "3. Start the application with 'go run main.go'"
