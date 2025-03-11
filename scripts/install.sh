#!/bin/bash

if [ -z "$1" ]; then
    echo "Please specify the domain name or IP address"
    exit 1
else
    DOMAIN="$1"
fi

echo "Installing the application"

# We should install the dependencies here such as docker

echo "Setting up environment variables"

cp .env.sample .env

perl -pi -e "s|DOMAIN|${DOMAIN}|g" .env

if docker compose up --build -d; then
    echo "Started successfully"
else
    echo "Failed to start"
    exit 1
fi

# once the application is up, we should create a default user, 
# ideally we will call the api here and register the user, if no users found on the table then register should work, else it should fail
# we will take the email in the second argument
# we will generate the random password and print it out for the user
# and also we should setup proxy for this application
# print out the application url for the user

echo "Application installed successfully"