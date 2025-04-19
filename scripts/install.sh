#!/bin/bash

if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (sudo)"
    exit 1
fi

if ! command -v python3 &> /dev/null; then
    echo "Python 3 is not installed. Please install Python 3 before running this script."
    exit 1
fi

TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

echo "Downloading Nixopus..."
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/installer

echo "Starting installation..."
python3 install.py

cd $TEMP_DIR/..
rm -rf $TEMP_DIR

echo "Installation completed!"