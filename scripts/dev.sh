#!/bin/bash

# Function to check if node_modules exists and is not empty
check_node_modules() {
  if [ -d "view/node_modules" ] && [ "$(ls -A view/node_modules 2>/dev/null)" ]; then
    return 0 
  else
    return 1
  fi
}

# Function to check if Go modules are already downloaded
check_go_modules() {
  if [ -d "api/vendor" ] || grep -q "require" "api/go.mod" && [ -d "$GOPATH/pkg/mod" ]; then
    if (cd api && go list -m all > /dev/null 2>&1); then
      return 0 
    fi
  fi
  return 1
}

# Start API server
(
  cd api
  if ! check_go_modules; then
    echo "Installing Go dependencies..."
    go mod tidy
  else
    echo "Go dependencies already installed, skipping..."
  fi
  air
) &

# Start frontend
(
  cd view
  if ! check_node_modules; then
    echo "Installing Node dependencies..."
    yarn install --frozen-lockfile
  else
    echo "Node dependencies already installed, skipping..."
  fi
  yarn dev
) &

wait