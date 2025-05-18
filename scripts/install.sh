#!/bin/bash

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo)"
    exit 1
fi

for cmd in python3 pip3 git; do
    if ! command -v $cmd &> /dev/null; then
        echo "$cmd is not installed. Please install $cmd before running this script."
        exit 1
    fi
done

ENV="production"
for arg in "$@"; do
    if [[ $arg == "--env"* ]]; then
        ENV=$(echo $arg | cut -d'=' -f2)
        break
    fi
done

if [ "$ENV" == "staging" ]; then
    NIXOPUS_DIR="/etc/nixopus-staging"
    SOURCE_DIR="$NIXOPUS_DIR/source"
    BRANCH="feat/installer"
else
    NIXOPUS_DIR="/etc/nixopus"
    SOURCE_DIR="$NIXOPUS_DIR/source"
    BRANCH="master"
fi

mkdir -p $NIXOPUS_DIR
mkdir -p $SOURCE_DIR

if [ -d "$SOURCE_DIR/.git" ]; then
    cd $SOURCE_DIR
    git fetch --all
    git reset --hard origin/$BRANCH
    git checkout $BRANCH
    git pull
else
    rm -rf $SOURCE_DIR/* $SOURCE_DIR/.[!.]*
    git clone https://github.com/raghavyuva/nixopus.git $SOURCE_DIR
    cd $SOURCE_DIR
    git checkout $BRANCH
fi

echo "Setting up Caddy configuration..."
rm -rf $NIXOPUS_DIR/caddy > /dev/null 2>&1
mkdir -p $NIXOPUS_DIR/caddy > /dev/null 2>&1
echo '{
	admin 0.0.0.0:2019
	log {
		format json
		level INFO
	}
}' > $NIXOPUS_DIR/caddy/Caddyfile

cd $SOURCE_DIR/installer
echo "Setting up Nixopus Installation Environment..."
python3 -m venv venv > /dev/null 2>&1
source venv/bin/activate > /dev/null 2>&1

echo "Installing dependencies..."
pip install --upgrade pip > /dev/null 2>&1
pip install -r requirements.txt > /dev/null 2>&1

echo "Starting Nixopus Installation..."
PYTHONPATH=$SOURCE_DIR/installer python3 install.py "$@"

deactivate > /dev/null 2>&1