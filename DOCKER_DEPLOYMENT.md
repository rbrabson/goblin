# Docker Setup and Deployment Guide

This guide covers building, deploying, and running the Goblin Discord bot using Docker.

## 📦 Building Docker Images

### Local Build

Build the image locally for testing:

```bash
# Build with default tag
docker build -platform=linux/amd64 -t goblin .

# Build with specific tag
docker build -platform=linux/amd64 -t goblin:v1.0.0 .

# Build with multiple tags
docker build -platform=linux/amd64 -t goblin:latest -t goblin:v1.0.0 .
```

### Build Arguments (if needed)

```bash
# Example with build arguments
docker build --build-arg GO_VERSION=1.25 -t goblin .
```

## 🏷️ Tagging Images

### Tag for Docker Hub

```bash
# Tag for Docker Hub (replace 'username' with your Docker Hub username)
docker tag goblin:latest username/goblin:latest
docker tag goblin:latest username/goblin:v1.0.0
```

### Tag for Private Registry

```bash
# Tag for private registry
docker tag goblin:latest registry.company.com/goblin:latest
docker tag goblin:latest registry.company.com/goblin:v1.0.0
```

## 🚀 Pushing Images

### Push to Docker Hub

```bash
# Login to Docker Hub
docker login

# Push specific version
docker push username/goblin:v1.0.0

# Push latest
docker push username/goblin:latest

# Push all tags
docker push username/goblin --all-tags
```

### Push to Private Registry

```bash
# Login to private registry
docker login registry.company.com

# Push images
docker push registry.company.com/goblin:latest
docker push registry.company.com/goblin:v1.0.0
```

## 📥 Pulling Images

### Pull from Docker Hub

```bash
# Pull latest version
docker pull username/goblin:latest

# Pull specific version
docker pull username/goblin:v1.0.0
```

### Pull from Private Registry

```bash
# Pull from private registry
docker pull registry.company.com/goblin:latest
```

## 🔧 Docker Compose Deployment

### Option 1: Using Pre-built Image with Environment Variables

Update `sample_docker_compose.yaml`:

```yaml
services:
  goblin_bot:
    container_name: goblin
    image: username/goblin:latest  # Replace with your image
    environment:
services:
  goblin:
    container_name: goblin-bot
    image: rbrabson/private:goblin
    environment:
      DISCORD_BOT_TOKEN: ${DISCORD_BOT_TOKEN}
      DISCORD_APP_ID: ${DISCORD_APP_ID}
      DISCORD_CONFIG_DIR: ${DISCORD_CONFIG_DIR}
      DISCORD_DEFAULT_THEME: ${DISCORD_DEFAULT_THEME}
      MONGODB_DATABASE: ${MONGODB_DATABASE}
      MONGODB_URI: ${MONGODB_URI}
      LOG_LEVEL: ${LOG_LEVEL}
    entrypoint: /goblin
    restart: unless-stopped

  mongodb:
    container_name: goblin-mongo
    image: mongo:latest
    environment:
      - MONGO_INITDB_ROOT_USERNAME=${ROOT_USERNAME}$
      - MONGO_INITDB_ROOT_PASSWORD=${ROOT_PASSWORD}
      - MONGO_INITDB_DATABASE=${DATABASE_NAME}
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    restart: unless-stopped

volumes:
  mongodb_data:
    driver: local
```

Environment variables used by the deployment include:

```env
# Discord Bot Configuration
DISCORD_BOT_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_app_id_here
DISCORD_CONFIG_DIR=/config
DISCORD_DEFAULT_THEME=clash
LOG_LEVEL=info

# MongoDB Configuration for the Discord bot
MONGODB_URI=mongodb+srv://user:password@server/database?retryWrites=true&w=majority

# Local MongoDB Container Settings
MONGODB_ROOT_USER=admin
MONGODB_ROOT_PASSWORD=your_mongo_password
MONGODB_DATABASE=goblin
```

### Option 2: Using .env File (Recommended)

1. **Create a `.env` file:**

    ```bash
    cp sample.env .env
    ```

2. **Edit `.env` with your values:**

    ```env
    # Discord Bot Configuration
    DISCORD_BOT_TOKEN=your_bot_token_here
    DISCORD_APP_ID=your_app_id_here
    DISCORD_CONFIG_DIR=/config
    DISCORD_DEFAULT_THEME=clash
    LOG_LEVEL=info

    # MongoDB Configuration for the Discord bot
    MONGODB_URI=mongodb+srv://user:password@server/database?retryWrites=true&w=majority

    # Local MongoDB Container Settings
    MONGODB_ROOT_USER=admin
    MONGODB_ROOT_PASSWORD=your_mongo_password
    MONGODB_DATABASE=goblin
    ```

3. **Update Docker Compose to use .env:**

    ```yaml
    services:
      goblin_bot:
        container_name: goblin
        image: username/goblin:latest  # Replace with your image
        env_file:
          - .env
        entrypoint: /goblin
        restart: unless-stopped
        volumes:
          - ./config:/config:ro  # Mount config directory (optional)

      mongodb:
        container_name: goblin-mongo
        image: mongo:latest
        env_file:
          - .env
        environment:
          - MONGO_INITDB_ROOT_USERNAME=${MONGODB_ROOT_USER}
          - MONGO_INITDB_ROOT_PASSWORD=${MONGODB_ROOT_PASSWORD}
          - MONGO_INITDB_DATABASE=${MONGODB_DATABASE}
        ports:
          - "27017:27017"
        volumes:
          - mongodb_data:/data/db
        restart: unless-stopped

    volumes:
      mongodb_data:
        driver: local
    ```

### Option 3: Building Image with Docker Compose

For development or custom builds:

```yaml
services:
  goblin_bot:
    container_name: goblin
    build:
      context: .
      dockerfile: ./Dockerfile
    env_file:
      - .env
    entrypoint: /goblin
    restart: unless-stopped
    depends_on:
      - mongodb

  mongodb:
    container_name: goblin-mongo
    image: mongo:latest
    env_file:
      - .env
    environment:
      - MONGO_INITDB_ROOT_USERNAME=${MONGODB_ROOT_USER}
      - MONGO_INITDB_ROOT_PASSWORD=${MONGODB_ROOT_PASSWORD}
      - MONGO_INITDB_DATABASE=${MONGODB_DATABASE}
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    restart: unless-stopped

volumes:
  mongodb_data:
    driver: local
```

## 🚀 Running the Application

### Setup Steps

1. **Copy sample files:**

   ```bash
   cp sample_docker_compose.yaml docker-compose.yaml
   cp sample.env .env  # If using .env option
   ```

2. **Edit configuration:**
   - Update `docker-compose.yaml` with your image name
   - Edit `.env` with your actual values (if using .env option)
   - Or edit environment variables directly in `docker-compose.yaml`

3. **Start services:**

   ```bash
   # Start in background
   docker-compose up -d
   
   # Start with logs visible
   docker-compose up
   
   # Build and start (if using build option)
   docker-compose up --build -d
   ```

### Management Commands

```bash
# View logs
docker-compose logs -f goblin_bot
docker-compose logs -f mongodb

# Stop services
docker-compose stop

# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes
docker-compose down -v

# Restart specific service
docker-compose restart goblin_bot

# Update and restart (pull new image)
docker-compose pull
docker-compose up -d

# View running services
docker-compose ps

# Execute commands in running container
docker-compose exec goblin_bot /bin/sh
```

## 🔒 Security Best Practices

1. **Use .env files for sensitive data**
2. **Never commit .env files to version control**
3. **Use specific image tags instead of 'latest' in production**
4. **Regularly update base images for security patches**
5. **Use secrets management for production deployments**

## 🐛 Troubleshooting

### Common Issues

**Container exits immediately:**

```bash
# Check logs
docker-compose logs goblin_bot

# Check if environment variables are set
docker-compose exec goblin_bot env
```

**MongoDB connection issues:**

```bash
# Check MongoDB logs
docker-compose logs mongodb

# Test MongoDB connection
docker-compose exec mongodb mongosh
```

**Image pull failures:**

```bash
# Login to registry
docker login

# Check image exists
docker pull username/goblin:latest
```

### Debug Mode

Run with debug logging:

```bash
# Set LOG_LEVEL=debug in .env or docker-compose.yaml
LOG_LEVEL=debug docker-compose up
```

## 📋 Environment Variables Reference

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DISCORD_BOT_TOKEN` | Discord bot token | Yes | - |
| `DISCORD_APP_ID` | Discord application ID | Yes | - |
| `DISCORD_CONFIG_DIR` | Configuration directory path | No | `/config` |
| `DISCORD_DEFAULT_THEME` | Default theme name | No | `clash` |
| `MONGODB_DATABASE` | Name of the MogoDB database | Yes | - |
| `MONGODB_URI` | MongoDB connection string | Yes | - |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | No | `info` |
| `DISCORD_GUILD_ID` | Development server ID (optional) | No | - |
| `MONGODB_ROOT_USER` | MongoDB root username (for local container) | No | `admin` |
| `MONGODB_ROOT_PASSWORD` | MongoDB root password (for local container) | Yes* | - |
| `MONGODB_DATABASE` | MongoDB database name (for local container) | No | `goblin` |

*Required when using local MongoDB container
