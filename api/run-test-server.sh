#!/bin/bash
export DATABASE_URL="postgresql://nixopus:nixopus@localhost:5433/nixopus_test?sslmode=disable"
export AUTH_SERVICE_URL="http://localhost:9090"
export AUTH_SERVICE_SECRET="test-secret-for-ci"
export REDIS_URL="redis://localhost:6379"
export PORT=8080
export ENV=test
export SECRET_MANAGER_ENABLED=false
export ALLOWED_ORIGIN="http://localhost:3000"
export CADDY_PORT=2019
export GITHUB_APP_WEBHOOK_SECRET="test"
export S3_ENDPOINT=""
export S3_BUCKET=""
export S3_REGION=""
export S3_ACCESS_KEY=""
export S3_SECRET_KEY=""
export AUTH_ISSUER="http://localhost:9090"
export AUTH_JWKS_URL="http://localhost:9090/api/auth/jwks"
export AGENT_URL="http://localhost:4117"
export OAUTH_CLIENT_ID="test-client"
export OAUTH_CLIENT_SECRET="test-secret"
export TRIAL_PERIOD_DAYS=7

cd /Users/shaale/nixopus/api
exec ./bin/nixopus-api
