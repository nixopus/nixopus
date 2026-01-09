UPDATE extensions
SET featured = false
WHERE extension_id IN ('deploy-uptime-kuma', 'deploy-n8n', 'deploy-excalidraw', 'deploy-minio', 'deploy-redis', 'deploy-postgres', 'deploy-ollama', 'deploy-qdrant', 'deploy-postiz', 'deploy-code-server', 'deploy-changedetection-io', 'deploy-dashy', 'deploy-freshrss', 'deploy-mastodon', 'fail2ban-ssh', 'deploy-webhook-tester')
  AND deleted_at IS NULL;

