#!/bin/bash

################################################################################
# Nixopus Development Setup Script
#
# Purpose: Start API and View services locally with hot reloading
# Prerequisites: Dependencies (DB, Redis, SuperTokens, Caddy) must be running
#                via docker-compose-dev.yml
#
# Usage: ./scripts/dev.sh
#        API_PORT=3000 VIEW_PORT=3001 ./scripts/dev.sh  # Custom ports
################################################################################

set -e

API_PORT=${API_PORT:-8080}
VIEW_PORT=${VIEW_PORT:-3000}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="${SCRIPT_DIR}/api"
VIEW_DIR="${SCRIPT_DIR}/view"
API_PID_FILE="${SCRIPTS_DIR}/.api.pid"
VIEW_PID_FILE="${SCRIPTS_DIR}/.view.pid"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color


cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    if [ -f "$API_PID_FILE" ]; then
        API_PID=$(cat "$API_PID_FILE")
        if kill -0 "$API_PID" 2>/dev/null || ps -p "$API_PID" > /dev/null 2>&1; then
            echo -e "${BLUE}Stopping API (PID: $API_PID)...${NC}"
            kill "$API_PID" 2>/dev/null || true
        fi
        rm -f "$API_PID_FILE"
    fi
    
    if [ -f "$VIEW_PID_FILE" ]; then
        VIEW_PID=$(cat "$VIEW_PID_FILE")
        if kill -0 "$VIEW_PID" 2>/dev/null || ps -p "$VIEW_PID" > /dev/null 2>&1; then
            echo -e "${BLUE}Stopping View (PID: $VIEW_PID)...${NC}"
            kill "$VIEW_PID" 2>/dev/null || true
        fi
        rm -f "$VIEW_PID_FILE"
    fi
    
    echo -e "${GREEN}Cleanup complete${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

detect_docker_compose() {
    if docker compose version &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker compose"
    elif docker-compose version &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker-compose"
    else
        echo -e "${RED}Docker Compose is not installed. Please install Docker Compose.${NC}"
        exit 1
    fi
}

check_dependencies() {
    echo -e "${BLUE}Checking if dependencies are running...${NC}"
    
    if ! docker ps 2>/dev/null | grep -q "nixopus-db\|nixopus-redis\|nixopus-supertokens"; then
        echo -e "${YELLOW}Warning: Dependencies may not be running.${NC}"
        echo -e "${YELLOW}Start them with: ${DOCKER_COMPOSE_CMD} -f docker-compose-dev.yml up -d${NC}"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✓ Dependencies are running${NC}"
    fi
}

check_air() {
    echo -e "${BLUE}Checking if air is installed...${NC}"
    
    if ! command -v air &> /dev/null; then
        echo -e "${YELLOW}air is not installed. Installing...${NC}"
        go install github.com/air-verse/air@latest
        
        if ! command -v air &> /dev/null; then
            echo -e "${RED}Failed to install air. Please install it manually:${NC}"
            echo -e "${RED}  go install github.com/air-verse/air@latest${NC}"
            exit 1
        fi
    fi
    
    echo -e "${GREEN}✓ air is installed${NC}"
}

check_node() {
    echo -e "${BLUE}Checking if Node.js is installed...${NC}"
    
    if ! command -v node &> /dev/null; then
        echo -e "${RED}Node.js is not installed. Please install Node.js first.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Node.js is installed ($(node --version))${NC}"
}

check_yarn() {
    echo -e "${BLUE}Checking if yarn is installed...${NC}"
    
    if ! command -v yarn &> /dev/null; then
        echo -e "${YELLOW}yarn is not installed. Installing...${NC}"
        npm install -g yarn
        
        if ! command -v yarn &> /dev/null; then
            echo -e "${RED}Failed to install yarn. Please install it manually:${NC}"
            echo -e "${RED}  npm install -g yarn${NC}"
            exit 1
        fi
    fi
    
    echo -e "${GREEN}✓ yarn is installed ($(yarn --version))${NC}"
}

setup_ssh_key() {
    local SSH_KEY_PATH="${HOME}/.ssh/id_rsa_nixopus"
    
    if [ ! -f "$SSH_KEY_PATH" ]; then
        echo -e "${YELLOW}SSH key not found at ${SSH_KEY_PATH}${NC}"
        read -p "Do you want to generate an SSH key? (y/N) " -n 1 -r
        echo
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}Generating SSH key...${NC}"
            mkdir -p "${HOME}/.ssh"
            ssh-keygen -t rsa -b 4096 -f "$SSH_KEY_PATH" -N "" -C "nixopus-dev"
            echo -e "${GREEN}✓ SSH key generated${NC}"
            
            if [ -f "${HOME}/.ssh/authorized_keys" ]; then
                if ! grep -q "$(cat ${SSH_KEY_PATH}.pub)" "${HOME}/.ssh/authorized_keys" 2>/dev/null; then
                    cat "${SSH_KEY_PATH}.pub" >> "${HOME}/.ssh/authorized_keys"
                    echo -e "${GREEN}✓ Added public key to authorized_keys${NC}"
                fi
            else
                cat "${SSH_KEY_PATH}.pub" > "${HOME}/.ssh/authorized_keys"
                chmod 600 "${HOME}/.ssh/authorized_keys"
                echo -e "${GREEN}✓ Created authorized_keys file${NC}"
            fi
        else
            echo -e "${YELLOW}You'll need to set SSH_PRIVATE_KEY manually in the .env file${NC}"
        fi
    else
        echo -e "${GREEN}✓ SSH key found${NC}"
    fi
}

create_api_env() {
    local API_ENV_FILE="${API_DIR}/.env"
    
    if [ -f "$API_ENV_FILE" ]; then
        echo -e "${YELLOW}  .env file already exists, skipping creation${NC}"
        return
    fi
    
    echo -e "${BLUE}Creating API .env file...${NC}"
    
    local DB_PORT=${DB_PORT:-5432}
    local DB_USERNAME=${USERNAME:-postgres}
    local DB_PASSWORD=${PASSWORD:-changeme}
    local DB_NAME=${DB_NAME:-postgres}
    local REDIS_PORT=${REDIS_PORT:-6379}
    local REDIS_URL=${REDIS_URL:-redis://localhost:${REDIS_PORT}}
    local SUPERTOKENS_PORT=${SUPERTOKENS_PORT:-3567}
    local CADDY_ADMIN_PORT=${CADDY_ADMIN_PORT:-2019}
    local SSH_USER=${SSH_USER:-${USER:-user}}
    local SSH_KEY_PATH="${HOME}/.ssh/id_rsa_nixopus"
    local MOUNT_PATH=${MOUNT_PATH:-./configs}
    local LOGS_PATH=${LOGS_PATH:-./logs}
    local APP_VERSION=${APP_VERSION:-0.1.0-alpha.11}
    
    local SSH_PASSWORD_VAR=${SSH_PASSWORD:-}
    if [ -z "$SSH_PASSWORD_VAR" ]; then
        echo -e "${BLUE}SSH Password (leave empty if using key-based auth):${NC}"
        read -s SSH_PASSWORD_VAR
        echo
    fi
    
    cat > "$API_ENV_FILE" << EOF
# API Configuration
PORT=${API_PORT}
API_PORT=${API_PORT}

# Database Configuration
HOST_NAME=localhost
DB_PORT=${DB_PORT}
USERNAME=${DB_USERNAME}
PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
SSL_MODE=disable

# Redis Configuration
REDIS_URL=${REDIS_URL}

# SSH Configuration
SSH_HOST=localhost
SSH_PORT=22
SSH_USER=${SSH_USER}
SSH_PRIVATE_KEY=${SSH_KEY_PATH}
SSH_PASSWORD=${SSH_PASSWORD_VAR}

# Docker Configuration
DOCKER_HOST=unix:///var/run/docker.sock
DOCKER_PORT=2376

# Caddy Configuration
CADDY_ENDPOINT=http://localhost:${CADDY_ADMIN_PORT}

# SuperTokens Configuration
SUPERTOKENS_API_KEY=NixopusSuperTokensAPIKey
SUPERTOKENS_API_DOMAIN=http://localhost:${API_PORT}
SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:${VIEW_PORT}
SUPERTOKENS_CONNECTION_URI=http://localhost:${SUPERTOKENS_PORT}

# CORS Configuration
ALLOWED_ORIGIN=http://localhost:${VIEW_PORT}

# Application Configuration
ENV=development
LOGS_PATH=${LOGS_PATH}
MOUNT_PATH=${MOUNT_PATH}
APP_VERSION=${APP_VERSION}
EOF
    
    echo -e "${GREEN}  ✓ Created ${API_ENV_FILE}${NC}"
}

create_view_env() {
    local VIEW_ENV_FILE="${VIEW_DIR}/.env"
    
    if [ -f "$VIEW_ENV_FILE" ]; then
        echo -e "${YELLOW}  .env file already exists, skipping creation${NC}"
        return
    fi
    
    echo -e "${BLUE}Creating View .env file...${NC}"
    
    local API_URL=${API_URL:-http://localhost:${API_PORT}/api}
    local WEBSOCKET_URL=${WEBSOCKET_URL:-ws://localhost:${API_PORT}/ws}
    local WEBHOOK_URL=${WEBHOOK_URL:-http://localhost:${API_PORT}/api/v1/webhook}
    local LOGS_PATH=${LOGS_PATH:-./logs}
    
    cat > "$VIEW_ENV_FILE" << EOF
# View Configuration
PORT=${VIEW_PORT}
NEXT_PUBLIC_PORT=${VIEW_PORT}

# API Configuration
NEXT_PUBLIC_API_URL=${API_URL}
API_URL=${API_URL}
WEBSOCKET_URL=${WEBSOCKET_URL}
WEBHOOK_URL=${WEBHOOK_URL}

# Domain Configuration
VIEW_DOMAIN=http://localhost:${VIEW_PORT}

# Environment
NODE_ENV=development
NEXT_TELEMETRY_DISABLED=1
LOGS_PATH=${LOGS_PATH}
EOF
    
    echo -e "${GREEN}  ✓ Created ${VIEW_ENV_FILE}${NC}"
}

setup_api() {
    echo -e "\n${BLUE}Setting up API...${NC}"
    
    if [ ! -d "$API_DIR" ]; then
        echo -e "${RED}API directory not found: $API_DIR${NC}"
        exit 1
    fi
    
    setup_ssh_key
    create_api_env
    
    cd "$API_DIR"
    if [ ! -d "vendor" ] && [ -f "go.mod" ]; then
        echo -e "${BLUE}Installing Go dependencies...${NC}"
        go mod download
    fi
    
    if [ ! -f ".air.toml" ]; then
        echo -e "${YELLOW}Warning: .air.toml not found. Air will use default config${NC}"
    fi
    
    echo -e "${GREEN}✓ API setup complete${NC}"
}

setup_view() {
    echo -e "\n${BLUE}Setting up View...${NC}"
    
    if [ ! -d "$VIEW_DIR" ]; then
        echo -e "${RED}View directory not found: $VIEW_DIR${NC}"
        exit 1
    fi
    
    create_view_env
    
    cd "$VIEW_DIR"
    if [ ! -d "node_modules" ]; then
        echo -e "${BLUE}Installing Node.js dependencies...${NC}"
        yarn install
    fi
    
    echo -e "${GREEN}✓ View setup complete${NC}"
}

start_api() {
    echo -e "\n${BLUE}Starting API on port $API_PORT...${NC}"
    
    cd "$API_DIR"
    
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
    fi
    
    export PORT="${PORT:-$API_PORT}"
    export API_PORT="${API_PORT:-$API_PORT}"
    
    air > "${SCRIPTS_DIR}/.api.log" 2>&1 &
    API_PID=$!
    echo "$API_PID" > "$API_PID_FILE"
    
    echo -e "${GREEN}✓ API started (PID: $API_PID)${NC}"
    echo -e "${BLUE}  Logs: tail -f ${SCRIPTS_DIR}/.api.log${NC}"
    echo -e "${BLUE}  URL:  http://localhost:${PORT}${NC}"
}

start_view() {
    echo -e "\n${BLUE}Starting View on port $VIEW_PORT...${NC}"
    
    cd "$VIEW_DIR"
    
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
    fi
    
    export PORT="${PORT:-$VIEW_PORT}"
    export NEXT_PUBLIC_PORT="${NEXT_PUBLIC_PORT:-$VIEW_PORT}"
    export NEXT_PUBLIC_API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:$API_PORT/api}"
    export NODE_ENV="${NODE_ENV:-development}"
    
    yarn dev --port "${PORT}" > "${SCRIPTS_DIR}/.view.log" 2>&1 &
    VIEW_PID=$!
    echo "$VIEW_PID" > "$VIEW_PID_FILE"
    
    echo -e "${GREEN}✓ View started (PID: $VIEW_PID)${NC}"
    echo -e "${BLUE}  Logs: tail -f ${SCRIPTS_DIR}/.view.log${NC}"
    echo -e "${BLUE}  URL:  http://localhost:${PORT}${NC}"
}

main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Nixopus Development Setup${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    detect_docker_compose
    check_dependencies
    check_air
    check_node
    check_yarn
    
    setup_api
    setup_view
    
    start_api
    sleep 2
    start_view
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Development servers are running!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}API:${NC}  http://localhost:$API_PORT"
    echo -e "${BLUE}View:${NC} http://localhost:$VIEW_PORT"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""
    
    wait
}

main
