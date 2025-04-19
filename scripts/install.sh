#!/bin/bash

if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (sudo)"
    exit 1
fi

if ! command -v python3 &> /dev/null; then
    echo "Python 3 is not installed. Please install Python 3 before running this script."
    exit 1
fi

if ! command -v pip3 &> /dev/null; then
    echo "pip3 is not installed. Please install pip3 before running this script."
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo "git is not installed. Please install git before running this script."
    exit 1
fi

TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

echo "Downloading Nixopus..."
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/installer

echo "Setting up Python virtual environment..."
python3 -m venv venv
source venv/bin/activate

echo "Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

echo "Starting installation..."
cat docker-compose.yml
python3 install.py "$@"

deactivate
cd $TEMP_DIR/..
rm -rf $TEMP_DIR

echo "Installation completed!"