#!/bin/bash

if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root" >&2
    echo "Please run with sudo or as root user" >&2
    exit 1
fi

KEY_TYPE="rsa"
KEY_NAME="nixopus"
APP_DIR="/opt/nixopus"
SSH_DIR="$APP_DIR/.ssh"
SERVER_HOST="192.168.43.96"
SERVER_USER="root"

while [ $# -gt 0 ]; do
  case "$1" in
    --app-dir=*)
      APP_DIR="${1#*=}"
      SSH_DIR="$APP_DIR/.ssh"
      shift
      ;;
    --key-name=*)
      KEY_NAME="${1#*=}"
      shift
      ;;
    --key-type=*)
      KEY_TYPE="${1#*=}"
      shift
      ;;
    --server=*)
      SERVER_HOST="${1#*=}"
      shift
      ;;
    --user=*)
      SERVER_USER="${1#*=}"
      shift
      ;;
    *)
      echo "Unknown parameter: $1" >&2
      exit 1
      ;;
  esac
done

PRIVATE_KEY_FILE="$SSH_DIR/${KEY_NAME}_private_key"
PUBLIC_KEY_FILE="$SSH_DIR/${KEY_NAME}_private_key.pub"
KNOWN_HOSTS_FILE="$SSH_DIR/known_hosts"

if [ ! -d "$APP_DIR" ]; then
  echo "Creating application directory $APP_DIR..."
  mkdir -p "$APP_DIR"
  chmod 755 "$APP_DIR"
fi

if [ ! -d "$SSH_DIR" ]; then
  echo "Creating SSH directory $SSH_DIR..."
  mkdir -p "$SSH_DIR"
  chmod 700 "$SSH_DIR"
fi

echo "Generating $KEY_TYPE SSH key pair for application in $APP_DIR..."
ssh-keygen -t "$KEY_TYPE" -f "$PRIVATE_KEY_FILE" -N ""

chmod 600 "$PRIVATE_KEY_FILE"
chmod 644 "$PUBLIC_KEY_FILE"

echo "Adding server's host key to known_hosts file..."
ssh-keyscan -H "$SERVER_HOST" > "$KNOWN_HOSTS_FILE" 2>/dev/null

if [ $? -eq 0 ]; then
  echo "Successfully added $SERVER_HOST to known_hosts file at $KNOWN_HOSTS_FILE"
  chmod 644 "$KNOWN_HOSTS_FILE"
else
  echo "Failed to retrieve host key from $SERVER_HOST" >&2
  echo "Please ensure the server is reachable and SSH is running" >&2
fi

echo -e "\nKeys have been saved to:"
echo "Private key: $PRIVATE_KEY_FILE"
echo "Public key: $PUBLIC_KEY_FILE"
echo "Known hosts: $KNOWN_HOSTS_FILE"

echo -e "\nWould you like to copy the public key to the server? (y/n)"
read -r answer
if [[ "$answer" =~ ^[Yy]$ ]]; then
  echo "Copying public key to $SERVER_USER@$SERVER_HOST..."
  cat "$PUBLIC_KEY_FILE" | ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_HOST" "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
  
  if [ $? -eq 0 ]; then
    echo "Successfully copied public key to server"
  else
    echo "Failed to copy public key to server" >&2
    echo "Please manually add the public key to the server's authorized_keys file" >&2
  fi
fi

if [ -n "$SUDO_USER" ]; then
  echo "Setting ownership to $SUDO_USER..."
  chown -R "$SUDO_USER" "$APP_DIR"
else
  echo "Running as root user, ownership remains with root"
fi

echo "Environment variables:"
echo "SSH_PRIVATE_KEY_PATH=$PRIVATE_KEY_FILE"
echo "---------------------------------------"