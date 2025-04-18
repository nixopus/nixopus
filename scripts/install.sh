#!/bin/bash

if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (sudo)"
    exit 1
fi

TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

echo "Downloading Nixopus..."
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/installer

echo "Installing dependencies..."
pip install -r requirements.txt

echo "Starting installation..."
python3 install.py

cd /
rm -rf $TEMP_DIR

echo "Installation completed!"