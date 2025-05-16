FROM ubuntu:22.04

# Install basic dependencies
RUN apt-get update && apt-get install -y \
    curl git bash unzip wget build-essential \
    software-properties-common gnupg lsb-release \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GO_VERSION=1.23.6
RUN curl -LO https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install Node.js + npm + TypeScript
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g typescript ts-node next

# Set up Go tools if needed
RUN go install golang.org/x/tools/gopls@latest

WORKDIR /workspace/nixopus
