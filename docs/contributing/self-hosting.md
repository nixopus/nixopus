# Contributing to Nixopus Self-Hosting

This guide provides detailed instructions for contributing to the self-hosting capabilities of Nixopus.

## Overview

Self-hosting contributions can involve:

- Improving the installation process
- Enhancing Docker configurations
- Adding support for new deployment environments
- Optimizing performance for self-hosted instances
- Improving backup and restoration processes
- Enhancing security for self-hosted deployments

## Understanding the Self-Host Architecture

Nixopus self-hosting uses a containerized architecture with the following components:

```
nixopus-api        # Backend API service
nixopus-db         # PostgreSQL database
nixopus-redis      # Redis for caching and pub/sub
nixopus-view       # Frontend Next.js application
nixopus-caddy      # Caddy web server for reverse proxy
```

## Setting Up for Self-Host Development

1. **Prerequisites**
   - Docker and Docker Compose
   - Python 3.8 or higher
   - Access to a Linux-based system (preferred)
   - Basic understanding of containerization

2. **Environment Setup**

   ```bash
   # Clone the repository
   git clone https://github.com/raghavyuva/nixopus.git
   cd nixopus
   
   # Create a development branch
   git checkout -b feature/self-host-improvement
   ```

## Key Files for Self-Hosting

```
/
├── docker-compose.yml             # Main Docker Compose configuration
├── docker-compose-staging.yml     # Staging environment configuration
├── installer/                     # Installation scripts
│   ├── install.py                 # Main installation script
│   ├── environment.py             # Environment setup
│   ├── service_manager.py         # Service management
│   ├── docker_setup.py            # Docker configuration
│   └── validation.py              # Validation utilities
├── scripts/
│   └── install.sh                 # Installation shell script
└── Makefile                       # Make targets for common operations
```

## Making Self-Hosting Improvements

### 1. Improving the Installation Process

When enhancing the installation process:

1. **Identify Pain Points**
   - Look for error-prone steps
   - Find areas where users struggle
   - Consider platform-specific issues

2. **Modify Installation Scripts**

   Example improvement to `install.py`:

   ```python
   # Add better error handling for network issues
   def check_connectivity(self):
       try:
           response = requests.get("https://registry.hub.docker.com/v2/", timeout=5)
           return response.status_code == 200
       except requests.RequestException:
           print("\033[31mError: Cannot connect to Docker Hub. Please check your internet connection.\033[0m")
           return False
   
   # Use in installer
   if not installer.check_connectivity():
       print("Aborting installation due to connectivity issues.")
       sys.exit(1)
   ```

3. **Update Documentation**
   - Reflect your changes in the self-hosting documentation
   - Add troubleshooting tips for common issues

### 2. Enhancing Docker Configurations

1. **Optimize Container Settings**

   Example improvements:

   ```yaml
   # Add health checks to more services
   service-name:
     image: service-image
     healthcheck:
       test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
       interval: 30s
       timeout: 10s
       retries: 3
       start_period: 40s
   
   # Add resource constraints
   service-name:
     image: service-image
     deploy:
       resources:
         limits:
           cpus: '0.50'
           memory: 512M
         reservations:
           cpus: '0.25'
           memory: 256M
   ```

2. **Improve Networking Security**

   ```yaml
   # Restrict exposed ports
   services:
     api:
       ports:
         # Only expose to localhost
         - "127.0.0.1:8443:8443"
   
   # Use custom networks with isolation
   networks:
     frontend-network:
       name: nixopus-frontend
     backend-network:
       name: nixopus-backend
       internal: true  # Not exposed to host network
   ```

3. **Enhance Volume Management**

   ```yaml
   # Use named volumes with better organization
   volumes:
     postgres-data:
       name: nixopus-postgres-data
     redis-data:
       name: nixopus-redis-data
     
   services:
     db:
       volumes:
         - postgres-data:/var/lib/postgresql/data
   ```

### 3. Supporting New Deployment Environments

1. **Add Environment-Specific Configurations**

   Create new templates for specific environments:

   ```python
   def generate_env_config(self, environment):
       """Generate environment-specific configuration."""
       if environment == "kubernetes":
           return self._generate_kubernetes_config()
       elif environment == "aws":
           return self._generate_aws_config()
       else:
           return self._generate_standard_config()
   ```

2. **Create Environment Detection Logic**

   ```python
   def detect_environment(self):
       """Detect the current environment."""
       if os.path.exists("/var/run/kubernetes"):
           return "kubernetes"
       elif os.environ.get("AWS_EXECUTION_ENV"):
           return "aws"
       elif os.environ.get("AZURE_EXTENSION_DIR"):
           return "azure"
       else:
           return "standard"
   ```

3. **Add Platform-Specific Instructions**

   For each new platform, create detailed instructions in the documentation.

### 4. Performance Optimization

1. **Database Tuning**

   Create a utility script for database optimization:

   ```python
   def optimize_postgres_for_hardware(self):
       """Optimize PostgreSQL configuration based on available hardware."""
       total_memory = self._get_system_memory_mb()
       
       # Calculate optimal settings
       shared_buffers = max(int(total_memory * 0.25), 128)
       effective_cache_size = max(int(total_memory * 0.75), 256)
       
       # Generate configuration
       config = [
           f"shared_buffers = {shared_buffers}MB",
           f"effective_cache_size = {effective_cache_size}MB",
           "work_mem = 4MB",
           "maintenance_work_mem = 64MB",
           "max_connections = 100"
       ]
       
       return "\n".join(config)
   ```

2. **Caching Strategy Improvements**

   Enhance the Redis configuration:

   ```yaml
   nixopus-redis:
     image: redis:7-alpine
     command: >
       redis-server
       --appendonly yes
       --maxmemory 256mb
       --maxmemory-policy allkeys-lru
       --save 900 1
       --save 300 10
       --save 60 10000
   ```

### 5. Backup and Restore Improvements

1. **Automated Backup System**

   Create a backup service in Docker Compose:

   ```yaml
   nixopus-backup:
     image: postgres:14-alpine
     volumes:
       - ${DB_VOLUME:-/etc/nixopus/db}:/source
       - ${BACKUP_VOLUME:-/etc/nixopus/backups}:/backups
     environment:
       - POSTGRES_USER=${USERNAME}
       - POSTGRES_PASSWORD=${PASSWORD}
       - POSTGRES_DB=${DB_NAME}
     command: >
       sh -c 'while true; do
         pg_dump -U ${USERNAME} -d ${DB_NAME} -h nixopus-db -F c -f /backups/nixopus_$$(date +%Y%m%d_%H%M%S).dump;
         find /backups -name "nixopus_*.dump" -type f -mtime +7 -delete;
         sleep 86400;
       done'
     networks:
       - nixopus-network
   ```

2. **Restoration Script Enhancement**

   ```python
   def restore_from_backup(self, backup_path):
       """Restore database from a backup file."""
       if not os.path.exists(backup_path):
           print(f"Error: Backup file {backup_path} not found.")
           return False
           
       try:
           print(f"Stopping services...")
           self._stop_services()
           
           print(f"Restoring database from {backup_path}...")
           container_name = "nixopus-db"
           command = [
               "docker", "exec", "-i", container_name,
               "pg_restore", "-U", "nixopus", "-d", "nixopus", 
               "--clean", "--if-exists", "--no-owner", "--no-privileges"
           ]
           
           with open(backup_path, 'rb') as f:
               subprocess.run(command, stdin=f, check=True)
               
           print(f"Restarting services...")
           self._start_services()
           
           return True
       except Exception as e:
           print(f"Error during restoration: {str(e)}")
           return False
   ```

### 6. Security Enhancements

1. **Secrets Management**

   Implement a secure secrets management solution:

   ```python
   from cryptography.fernet import Fernet
   
   class SecretsManager:
       def __init__(self, key_path):
           self.key_path = key_path
           self._ensure_key()
           
       def _ensure_key(self):
           """Ensure encryption key exists."""
           if not os.path.exists(self.key_path):
               key = Fernet.generate_key()
               os.makedirs(os.path.dirname(self.key_path), exist_ok=True)
               with open(self.key_path, 'wb') as f:
                   f.write(key)
               os.chmod(self.key_path, 0o600)
           
       def encrypt(self, value):
           """Encrypt a value."""
           with open(self.key_path, 'rb') as f:
               key = f.read()
           cipher = Fernet(key)
           return cipher.encrypt(value.encode()).decode()
           
       def decrypt(self, value):
           """Decrypt a value."""
           with open(self.key_path, 'rb') as f:
               key = f.read()
           cipher = Fernet(key)
           return cipher.decrypt(value.encode()).decode()
   ```

2. **TLS Configuration for Services**

   Enhance the Caddy configuration for better security:

   ```json
   {
     "admin": {
       "listen": "127.0.0.1:2019"  # Only listen on localhost
     },
     "logging": {
       "logs": {
         "default": {
           "level": "ERROR"
         }
       }
     },
     "storage": {
       "module": "file_system",
       "root": "/data"
     },
     "apps": {
       "tls": {
         "certificates": {
           "automate": ["your-domain.com"]
         },
         "automation": {
           "policies": [{
             "issuer": {
               "module": "acme",
               "challenges": {
                 "http": {
                   "disabled": false
                 },
                 "tlsalpn": {
                   "disabled": true
                 }
               }
             }
           }]
         }
       },
       "http": {
         "servers": {
           "main": {
             "listen": [":443"],
             "routes": [
               {
                 "match": [{
                   "host": ["your-domain.com"]
                 }],
                 "handle": [{
                   "handler": "reverse_proxy",
                   "upstreams": [{
                     "dial": "nixopus-api:8443"
                   }]
                 }]
               }
             ]
           }
         }
       }
     }
   }
   ```

## Testing Self-Hosting Changes

1. **Test Installation on Multiple Platforms**
   - Ubuntu/Debian
   - CentOS/RHEL
   - Docker Desktop (Windows/macOS)

2. **Verify Upgrade Path**
   - Test upgrading from previous version
   - Ensure configuration is preserved
   - Validate data persistence

3. **Resource Utilization Testing**
   - Monitor memory usage
   - Track CPU utilization
   - Measure disk I/O
   - Test with different resource constraints

4. **Security Testing**
   - Verify encryption settings
   - Test access controls
   - Check for exposed ports/services
   - Validate TLS configuration

## Submitting Self-Hosting Contributions

1. **Document Your Changes**
   - Explain the purpose of your changes
   - Provide example configurations
   - Include testing results

2. **Create Pull Request**

   ```bash
   git add .
   git commit -m "feat(self-host): improve database backup automation"
   git push origin feature/self-host-improvement
   ```

3. **Complete PR Template**
   - Include problem description
   - Explain your solution
   - List tested environments
   - Note any limitations

4. **Run Integration Tests**
   - Ensure all services start correctly
   - Verify functionality with your changes
   - Test error scenarios
   - Validate performance implications

## Need Help?

If you need assistance with self-hosting contributions:

- Join the #self-hosting channel on Discord
- Create an issue for specific problems
- Ask in the community forums

Thank you for improving Nixopus self-hosting!
