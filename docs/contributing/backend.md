# Contributing to Nixopus Backend

This guide provides detailed instructions for contributing to the Nixopus backend codebase.

## Setup for Backend Development

1. **Prerequisites**
   - Go version 1.23.6 or newer
   - PostgreSQL
   - Docker and Docker Compose (recommended)

2. **Environment Setup**

   ```bash
   # Clone the repository
   git clone https://github.com/raghavyuva/nixopus.git
   cd nixopus
   
   # Set up PostgreSQL database
   createdb nixopus -U postgres
   createdb nixopus_test -U postgres
   
   # Copy environment template
   # Note: Be sure to update the environment variables to suit your setup.
   cp api/.env.sample api/.env
   
   # Configure SuperTokens (required for authentication)
   # Update the SuperTokens configuration in your .env file:
   # SUPERTOKENS_API_KEY=your-secure-api-key
   # SUPERTOKENS_API_DOMAIN=http://localhost:3567
   # SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:3000
   # SUPERTOKENS_CONNECTION_URI=http://localhost:3567
   
   # Install dependencies
   cd api
   go mod download
   ```

3. **SuperTokens Authentication Setup**

   Nixopus uses SuperTokens for authentication. You'll need to set up SuperTokens Core for development:

   ```bash
   # Install SuperTokens Core (using Docker)
   docker run -p 3567:3567 -d \
     --name supertokens-core \
     registry.supertokens.io/supertokens/supertokens-postgresql
   
   # Or install locally (see SuperTokens documentation)
   npm install -g supertokens
   supertokens start
   ```

   **Required Environment Variables:**
   ```bash
   # In your api/.env file
   SUPERTOKENS_API_KEY=NixopusSuperTokensAPIKey
   SUPERTOKENS_API_DOMAIN=http://localhost:3567
   SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:3000
   SUPERTOKENS_CONNECTION_URI=http://localhost:3567
   ```

   **For Production:**
   - Generate a secure random string for `SUPERTOKENS_API_KEY`
   - Update domains to match your production URLs
   - Ensure SuperTokens Core is accessible from your application

4. **Database Migrations**

Currently **the migration works automatically when starting the server**. However, you can run migrations manually using the following command:

   ```bash
   # Run migrations
   cd api
   go run migrations/main.go
   ```

4. **Loading Development Fixtures**

The project includes a comprehensive fixtures system for development and testing. You can load sample data using the following commands:

   ```bash
   cd api
   
   # Load fixtures without affecting existing data
   make fixtures-load
   
   # Drop and recreate all tables, then load fixtures (clean slate)
   make fixtures-recreate
   
   # Truncate all tables, then load fixtures
   make fixtures-clean
   
   # Get help on fixtures commands
   make fixtures-help
   ```

   **Available Fixture Files:**
   - `fixtures/development/complete.yml` - Loads all fixtures (uses imports)
   - `fixtures/development/users.yml` - User data only
   - `fixtures/development/organizations.yml` - Organization data only
   - `fixtures/development/roles.yml` - Role data only
   - `fixtures/development/permissions.yml` - Permission data only
   - `fixtures/development/role_permissions.yml` - Role-permission mappings
   - `fixtures/development/feature_flags.yml` - Feature flags
   - `fixtures/development/organization_users.yml` - User-organization relationships

   The `complete.yml` file uses import statements to load all individual files, making it easy to get a full development environment set up quickly.

*Note: [air](https://github.com/air-verse/air) as a dev-dependency so you can start the backend with the air command.*

## Project Structure

The backend follows a clean architecture approach:

```
api/
├── internal/
│   ├── features/      # Feature modules
│   ├── middleware/    # HTTP middleware
│   ├── config/        # Application configuration
│   ├── storage/       # Data storage implementation
│   ├── utils/         # Utility functions
│   └── types/         # Type definitions
├── migrations/        # Database migrations
└── tests/             # Test utilities
```

## Adding a New Feature

1. **Create a New Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Implement Your Feature**

   Create a new directory under `api/internal/features/` with the following structure:

   ```
   api/internal/features/your-feature/
   ├── controller/init.go   # HTTP handlers
   ├── service/service_name.go      # Business logic
   ├── storage/dao_name.go      # Data access
   └── types/type_name.go        # Type definitions
   ```

   Here's a sample implementation:

   ```go
   // types.go
   package yourfeature
   
   type YourEntity struct {
       ID        string `json:"id" db:"id"`
       Name      string `json:"name" db:"name"`
       CreatedAt string `json:"created_at" db:"created_at"`
       UpdatedAt string `json:"updated_at" db:"updated_at"`
   }
   
   // init.go (Controller)
   package yourfeature

   import "github.com/go-fuego/fuego"

   type Controller struct {
        service *Service
   }

   func NewController(s *Service)*Controller {
        return &Controller{service: s}
   }

   func (c *Controller) GetEntities(ctx fuego.Context) error {
        entities, err := c.service.ListEntities()
        if err != nil {
            return ctx.JSON(500, map[string]string{"error": err.Error()})
        }
        return ctx.JSON(200, entities)
   }

   func (c *Controller) CreateEntity(ctx fuego.Context) error {
        var input YourEntity
        if err := ctx.Bind(&input); err != nil {
            return ctx.JSON(400, map[string]string{"error": "invalid input"})
        }
        created, err := c.service.CreateEntity(&input)
        if err != nil {
            return ctx.JSON(500, map[string]string{"error": err.Error()})
        }
        return ctx.JSON(201, created)
   }

   // service.go or service_name.go
   package yourfeature

   type Service struct {
       storage *Storage
   }

   func NewService(storage *Storage)*Service {
       return &Service{storage: storage}
   }

   // init.go or storage.go
   package yourfeature

   import (
        "context"
        "github.com/uptrace/bun"
   )

   type Storage struct {
       DB *bun.DB
       Ctx context.Context
   }

   func NewFeatureStorage(db *bun.DB, ctx context.Context)*NewFeatureStorage {
       return &FeatureStorage{
            DB:  db,
            Ctx: ctx
        }
   }

   ```

3. **Register Routes**

   Update `api/internal/routes.go` to include your new endpoints:

   ```go
   // Register your feature routes
   yourFeatureStorage := yourfeature.NewStorage(db)
   yourFeatureService := yourfeature.NewService(yourFeatureStorage)
   yourFeatureController := yourfeature.NewController(yourFeatureService)
   
   api := router.Group("/api")
   {
       // Your feature endpoints
       featureGroup := api.Group("/your-feature")
       {
           featureGroup.GET("/", middleware.Authorize(), yourFeatureController.GetEntities)
           featureGroup.POST("/", middleware.Authorize(), yourFeatureController.CreateEntity)
           // Add more routes as needed
       }
   }
   ```

4. **Add Database Migrations**

   Create migration files in `api/migrations/your-feature/`:

   ```sql
   -- seq_number_create_your_feature_table.up.sql
   CREATE TABLE your_feature (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name TEXT NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
   );
   
   -- seq_number_create_your_feature_table.down.sql
   DROP TABLE IF EXISTS your_feature;
   ```

5. **Write Tests**

   Organize your tests in the `tests/` using a separate folder named after each feature:

   ```go
   // controller_test.go
   package yourfeature
   
   import (
       "testing"
       // Import other necessary packages
   )
   
   func TestGetEntity(t *testing.T) {
       // Test implementation
   }
   
   // service_test.go
   // storage_test.go
   ```

6. **Update API Documentation**

   Note that the docs will be updated automatically; the OpenAPI specification in `api/doc/openapi.json` will be updated automatically.

## Testing

1. **Run Unit Tests**

   ```bash
   cd api
   go test ./internal/features/your-feature/...
   ```

2. **Run Integration Tests**

   ```bash
   cd api
   go test ./api/internal/tests/... -tags=integration
   ```

3. **Run All Tests**

   ```bash
   cd api
   make test
   ```

## Optimizing Performance

1. **Use Database Indices** for frequently queried columns
2. **Implement Caching** for expensive operations
3. **Optimize SQL Queries** for better performance
4. **Add Proper Error Handling** and logging

## Code Style and Guidelines

1. Follow Go's [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
2. Use meaningful variable and function names
3. Add comments for complex logic
4. Structure code for readability and maintainability
5. Follow the existing project patterns

## Common Pitfalls

1. Forgetting to update migrations
2. Not handling database transactions properly
3. Missing error handling
4. Inadequate test coverage
5. Not considering performance implications

## Submitting Your Contribution

1. **Commit Changes**

   ```bash
   git add .
   git commit -m "feat: add your feature"
   ```

2. **Push and Create a Pull Request**

   ```bash
   git push origin feature/your-feature-name
   ```

3. Follow the PR template and provide detailed information about your changes.

## Need Help?

If you need assistance, feel free to:

- Create an issue on GitHub
- Reach out on the project's Discord channel
- Contact the maintainers directly

Thank you for contributing to Nixopus!
