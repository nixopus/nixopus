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
cd $TEMP_DIR > /dev/null 2>&1

echo "Downloading Nixopus..."
git clone https://github.com/raghavyuva/nixopus.git > /dev/null 2>&1
cd nixopus > /dev/null 2>&1

ENV="production"
for arg in "$@"; do
    if [[ $arg == "--env"* ]]; then
        ENV=$(echo $arg | cut -d'=' -f2)
        break
    fi
done

if [ "$ENV" == "staging" ]; then
    echo "Checking out to feat/develop branch for staging environment..."
    git checkout feat/develop > /dev/null 2>&1
    NIXOPUS_DIR="/etc/nixopus-staging"
else
    NIXOPUS_DIR="/etc/nixopus"
fi

cd installer > /dev/null 2>&1

echo "Setting up Nixopus Installation Environment..."
python3 -m venv venv > /dev/null 2>&1
source venv/bin/activate > /dev/null 2>&1

echo "Installing dependencies..."
pip install --upgrade pip > /dev/null 2>&1
pip install -r requirements.txt > /dev/null 2>&1

rm -rf $NIXOPUS_DIR/caddy > /dev/null 2>&1
mkdir -p $NIXOPUS_DIR/caddy > /dev/null 2>&1
echo '{
	admin 0.0.0.0:2019
	log {
		format json
		level INFO
	}
}' > $NIXOPUS_DIR/caddy/Caddyfile

echo "Starting Nixopus Installation..."
python3 install.py "$@"

deactivate > /dev/null 2>&1
cd $TEMP_DIR/.. > /dev/null 2>&1
rm -rf $TEMP_DIR > /dev/null 2>&1