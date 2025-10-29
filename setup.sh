#!/bin/bash

# Goblin Discord Bot - Quick Setup Script
# This script helps you set up the Docker configuration quickly

set -e

echo "🤖 Goblin Discord Bot - Docker Setup"
echo "===================================="
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "✅ Docker and Docker Compose are available"
echo

# Copy sample files if they don't exist
if [ ! -f "docker-compose.yaml" ]; then
    if [ -f "sample_docker_compose.yaml" ]; then
        cp sample_docker_compose.yaml docker-compose.yaml
        echo "✅ Created docker-compose.yaml from sample"
    else
        echo "❌ sample_docker_compose.yaml not found"
        exit 1
    fi
else
    echo "ℹ️  docker-compose.yaml already exists"
fi

if [ ! -f ".env" ]; then
    if [ -f "sample.env" ]; then
        cp sample.env .env
        echo "✅ Created .env from sample"
    else
        echo "❌ sample.env not found"
        exit 1
    fi
else
    echo "ℹ️  .env already exists"
fi

echo
echo "📝 Next steps:"
echo "1. Edit .env with your Discord bot token and other configuration"
echo "2. Edit docker-compose.yaml to choose your deployment option"
echo "3. Run: docker-compose up -d"
echo
echo "💡 See DOCKER_DEPLOYMENT.md for detailed instructions"
echo
echo "🚀 Quick start:"
echo "   nano .env                    # Edit configuration"
echo "   docker-compose up -d         # Start services"
echo "   docker-compose logs -f       # View logs"