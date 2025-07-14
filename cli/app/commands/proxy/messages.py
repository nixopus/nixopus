# Dry run messages
dry_run_mode = "🔍 DRY RUN MODE"
dry_run_command_would_be_executed = "The following command would be executed:"
dry_run_command = "Command:"
dry_run_port = "Port:"
dry_run_config_file = "Config file:"
end_dry_run = "--- End of dry run ---"

# Success messages
proxy_initialized_successfully = "Caddy proxy initialized successfully on port {port}"
proxy_status_running = "Caddy proxy is running on port {port}"
proxy_reloaded_successfully = "Caddy proxy configuration reloaded successfully on port {port}"
proxy_stopped_successfully = "Caddy proxy stopped successfully on port {port}"

# Error messages
proxy_init_failed = "Failed to initialize Caddy proxy"
proxy_status_stopped = "Caddy proxy is not running on port {port}"
proxy_status_failed = "Failed to check Caddy proxy status"
proxy_reload_failed = "Failed to reload Caddy proxy configuration"
proxy_stop_failed = "Failed to stop Caddy proxy"

# Validation messages
config_file_required = "Configuration file is required"
config_file_not_found = "Configuration file not found: {file}"
invalid_json_config = "Invalid JSON in configuration file: {error}"

# Connection messages
caddy_connection_failed = "Failed to connect to Caddy: {error}"
caddy_status_code_error = "Caddy returned status code: {code}"
caddy_load_failed = "Failed to load configuration: {code} - {response}"

# Debug messages
debug_init_proxy = "Initializing Caddy proxy on port: {port}"
debug_check_status = "Checking Caddy proxy status on port: {port}"
debug_reload_config = "Reloading Caddy proxy configuration on port: {port}"
debug_stop_proxy = "Stopping Caddy proxy on port: {port}"

# Info messages
info_caddy_running = "Caddy is running"
info_config_loaded = "Configuration loaded successfully"
info_caddy_stopped = "Caddy stopped successfully" 