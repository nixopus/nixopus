UPDATE extensions
SET featured = false
WHERE extension_id IN ('deploy-uptime-kuma', 'deploy-n8n', 'deploy-excalidraw', 'deploy-minio', 'deploy-redis', 'deploy-postgres', 'deploy-ollama')
  AND deleted_at IS NULL;

