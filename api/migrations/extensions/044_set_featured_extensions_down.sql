UPDATE extensions
SET featured = false
WHERE extension_id IN ('deploy-uptime-kuma', 'deploy-n8n', 'deploy-excalidraw', 'deploy-minio', 'deploy-redis', 'deploy-postgres', 'deploy-ollama', 'deploy-mastodon', 'deploy-code-server', 'fail2ban-ssh', 'deploy-changedetection-io', 'deploy-dashy', 'deploy-postiz', 'deploy-freshrss')
  AND deleted_at IS NULL;
