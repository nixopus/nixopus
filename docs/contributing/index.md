# Contributing to Nixopus

Thank you for your interest in contributing to Nixopus! This guide will help you get started with development and explain how to contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Documentation](#documentation)
- [Testing](#testing)
- [Docker & Self-Hosting](#docker--self-hosting)
- [Development Fixtures](#development-fixtures)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

Before contributing, please review and agree to our [Code of Conduct](/code-of-conduct/index.md). We're committed to maintaining a welcoming and inclusive community.

## Development Setup

### Prerequisites

- **Backend**: Go 1.23.6 or newer, PostgreSQL
- **Frontend**: Node.js 18.0 or higher, Yarn
- **Documentation**: Node.js 18.0 or higher, Yarn
- Docker and Docker Compose (recommended)

### Initial Setup

1. **Fork and Clone**
   ```bash
   git clone git@github.com:your_username/nixopus.git
   cd nixopus
   ```

2. **Set Up PostgreSQL Databases**
   ```bash
   createdb nixopus -U postgres
   createdb nixopus_test -U postgres
   ```

3. **Configure Backend Environment**
   ```bash
   cd api
   cp .env.sample .env
   # Update .env with your database credentials
   # Configure SuperTokens:
   # SUPERTOKENS_API_KEY=your-secure-api-key
   # SUPERTOKENS_API_DOMAIN=http://localhost:3567
   # SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:3000
   # SUPERTOKENS_CONNECTION_URI=http://localhost:3567
   ```

4. **Set Up SuperTokens Core** (required for authentication)
   ```bash
   # Using Docker (recommended)
   docker run -p 3567:3567 -d \
     --name supertokens-core \
     registry.supertokens.io/supertokens/supertokens-postgresql
   
   # Or install locally
   npm install -g supertokens
   supertokens start
   ```

5. **Install Dependencies**
   ```bash
   # Backend
   cd api
   go mod download
   
   # Frontend
   cd ../view
   yarn install
   
   # Documentation
   cd ../docs
   yarn install
   ```

6. **Load Development Fixtures** (optional but recommended)
   ```bash
   cd ../api
   make fixtures-load        # Load without affecting existing data
   # or
   make fixtures-recreate    # Clean slate (drops and recreates tables)
   ```

### Running the Application

1. **Start the API service**
   ```bash
   cd api
   air  # Uses air for hot reloading
   ```

2. **Start the Frontend**
   ```bash
   cd view
   yarn dev
   ```

3. **Start Documentation** (optional)
   ```bash
   cd docs
   yarn dev
   ```

The API runs on `http://localhost:8443` and the frontend on `http://localhost:3000`.

## Project Structure

### Backend Structure

```
api/
├── internal/
│   ├── features/      # Feature modules (auth, container, deploy, etc.)
│   ├── middleware/    # HTTP middleware
│   ├── config/        # Application configuration
│   ├── storage/       # Data storage implementation
│   └── utils/         # Utility functions
├── migrations/        # Database migrations
└── tests/             # Test utilities
```

### Frontend Structure

```
view/
├── app/               # Next.js App Router pages
├── components/        # React components
│   ├── ui/            # UI components (Shadcn)
│   ├── layout/        # Layout components
│   └── features/      # Feature-specific components
├── hooks/             # Custom React hooks
├── redux/             # Redux store and services
└── types/             # TypeScript type definitions
```

## Backend Development

### Adding a New Feature

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Create Feature Structure**
   Create a new directory under `api/internal/features/`:
   ```
   api/internal/features/your-feature/
   ├── controller/init.go   # HTTP handlers
   ├── service/service.go   # Business logic
   ├── storage/storage.go   # Data access
   └── types/types.go       # Type definitions
   ```

3. **Example Implementation**

   **types.go:**
   ```go
   package yourfeature
   
   type YourEntity struct {
       ID        string `json:"id" db:"id"`
       Name      string `json:"name" db:"name"`
       CreatedAt string `json:"created_at" db:"created_at"`
   }
   ```

   **controller/init.go:**
   ```go
   package yourfeature
   
   import "github.com/go-fuego/fuego"
   
   type Controller struct {
       service *Service
   }
   
   func NewController(s *Service) *Controller {
       return &Controller{service: s}
   }
   
   func (c *Controller) GetEntities(ctx fuego.Context) error {
       entities, err := c.service.ListEntities()
       if err != nil {
           return ctx.JSON(500, map[string]string{"error": err.Error()})
       }
       return ctx.JSON(200, entities)
   }
   ```

   **service/service.go:**
   ```go
   package yourfeature
   
   type Service struct {
       storage *Storage
   }
   
   func NewService(storage *Storage) *Service {
       return &Service{storage: storage}
   }
   ```

   **storage/storage.go:**
   ```go
   package yourfeature
   
   import (
       "context"
       "github.com/uptrace/bun"
   )
   
   type Storage struct {
       DB  *bun.DB
       Ctx context.Context
   }
   
   func NewStorage(db *bun.DB, ctx context.Context) *Storage {
       return &Storage{DB: db, Ctx: ctx}
   }
   ```

4. **Register Routes**
   Update `api/internal/routes.go`:
   ```go
   yourFeatureStorage := yourfeature.NewStorage(db, ctx)
   yourFeatureService := yourfeature.NewService(yourFeatureStorage)
   yourFeatureController := yourfeature.NewController(yourFeatureService)
   
   api := router.Group("/api")
   {
       featureGroup := api.Group("/your-feature")
       {
           featureGroup.GET("/", middleware.Authorize(), yourFeatureController.GetEntities)
       }
   }
   ```

5. **Add Database Migrations**
   Create migration files in `api/migrations/your-feature/`:
   ```sql
   -- seq_number_create_your_feature_table.up.sql
   CREATE TABLE your_feature (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name TEXT NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
   );
   
   -- seq_number_create_your_feature_table.down.sql
   DROP TABLE IF EXISTS your_feature;
   ```

   Migrations run automatically when starting the server, or manually:
   ```bash
   go run migrations/main.go
   ```

### Backend Testing

```bash
cd api
go test ./internal/features/your-feature/...
make test  # Run all tests
```

## Frontend Development

### Adding a New Feature

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Create a New Page** (if needed)
   ```
   view/app/your-feature/
   ├── page.tsx        # Main page component
   └── components/     # Page-specific components
   ```

   **Example page.tsx:**
   ```tsx
   'use client'
   
   import { useEffect } from 'react'
   import { useDispatch, useSelector } from 'react-redux'
   import { fetchYourFeatureData } from '@/redux/your-feature/actions'
   
   export default function YourFeaturePage() {
     const dispatch = useDispatch()
     const { data, loading, error } = useSelector(state => state.yourFeature)
     
     useEffect(() => {
       dispatch(fetchYourFeatureData())
     }, [dispatch])
     
     if (loading) return <div>Loading...</div>
     if (error) return <div>Error: {error}</div>
     
     return (
       <div className="container mx-auto py-8">
         <h1 className="text-2xl font-bold mb-4">Your Feature</h1>
       </div>
     )
   }
   ```

3. **Create Redux API Service**
   ```tsx
   // view/redux/services/your-feature/yourFeatureApi.ts
   import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
   import { baseQueryWithReauth } from '@/redux/base-query'
   
   export const yourFeatureApi = createApi({
     reducerPath: 'yourFeatureApi',
     baseQuery: baseQueryWithReauth,
     tagTypes: ['YourFeature'],
     endpoints: (builder) => ({
       getYourFeatures: builder.query<YourFeature[], void>({
         query: () => ({ url: '/api/your-feature', method: 'GET' }),
         providesTags: ['YourFeature'],
       }),
       createYourFeature: builder.mutation<YourFeature, CreateDto>({
         query: (body) => ({
           url: '/api/your-feature',
           method: 'POST',
           body,
         }),
         invalidatesTags: ['YourFeature'],
       }),
     }),
   })
   
   export const {
     useGetYourFeaturesQuery,
     useCreateYourFeatureMutation,
   } = yourFeatureApi
   ```

4. **Create Components**
   Use existing Shadcn UI components from `components/ui/` for consistency. Follow the design system and ensure accessibility.

### Frontend Testing

```bash
cd view
yarn lint          # Check for lint errors
yarn lint:fix      # Fix lint errors
yarn test          # Run tests
yarn build         # Build for production
```

### UI Guidelines

- Use Shadcn UI components from `components/ui/`
- Follow Tailwind CSS patterns
- Ensure responsive design
- Maintain accessibility standards
- Use TypeScript for type safety

## Documentation

### Documentation Structure

```
docs/
├── index.md              # Documentation homepage
├── architecture/         # System architecture
├── install/              # Installation instructions
├── features/             # Feature documentation
└── contributing/         # This file
```

### Adding Documentation

1. **Create a new markdown file**
   - For a new section: `docs/your-section/index.md`
   - For a new topic: `docs/existing-section/your-topic.md`

2. **Add frontmatter**
   ```markdown
   ---
   title: "Your Page Title"
   description: "A brief description"
   ---
   ```

3. **Follow Documentation Standards**
   - Use clear headers (`#`, `##`, `###`)
   - Include code blocks with syntax highlighting
   - Add tables for structured data
   - Use links for navigation
   - Store images in `/docs/public/`

4. **Update Navigation**
   Update `docs/.vitepress/config.mts` to add your page to the sidebar.

### Running Documentation Locally

```bash
cd docs
yarn dev
# Server runs at http://localhost:3001
```

## Testing

### Backend Tests

```bash
cd api
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -race ./...        # Race detection
make test                  # Run full test suite
```

### Frontend Tests

```bash
cd view
yarn test                  # Run tests
yarn test:watch            # Watch mode
yarn lint                  # Lint check
```

## Docker & Self-Hosting

### Docker Development

Nixopus uses Docker for containerization. Key files:
- `api/Dockerfile` - API service
- `view/Dockerfile` - Frontend service
- `docker-compose.yml` - Main compose file

### Best Practices

1. **Image Optimization**
   - Use multi-stage builds
   - Minimize layer size
   - Use specific image versions
   - Clean up in the same layer

2. **Security**
   - Use non-root users
   - Scan for vulnerabilities
   - Use minimal base images
   - Don't embed secrets

3. **Testing Docker Changes**
   ```bash
   docker-compose build
   docker-compose up -d
   docker-compose logs -f
   ```

### Self-Hosting Improvements

Key files for self-hosting:
- `installer/install.py` - Main installation script
- `docker-compose.yml` - Container configuration
- `scripts/install.sh` - Installation shell script

When improving self-hosting:
- Test on multiple platforms (Ubuntu, Debian, CentOS)
- Verify upgrade paths
- Test resource constraints
- Validate security settings

## Development Fixtures

Fixtures provide sample data for development. Available commands:

```bash
cd api
make fixtures-load        # Load fixtures (preserves existing data)
make fixtures-recreate    # Clean slate (drops and recreates tables)
make fixtures-clean       # Truncate tables and reload
make fixtures-help        # Show help
```

### Fixture Files

```
api/fixtures/development/
├── complete.yml              # Main entry (imports all)
├── users.yml                 # User accounts
├── organizations.yml         # Organizations
├── roles.yml                 # Roles
├── permissions.yml           # Permissions
├── role_permissions.yml      # Role-permission mappings
├── feature_flags.yml         # Feature flags
└── organization_users.yml    # User-organization relationships
```

### Adding New Fixtures

1. Create a YAML file in `api/fixtures/development/`:
   ```yaml
   - table: my_feature
     data:
       - id: "550e8400-e29b-41d4-a716-446655440001"
         name: "Sample Feature"
         created_at: "2024-01-01T00:00:00Z"
   ```

2. Add to `complete.yml`:
   ```yaml
   - import: my_feature.yml
   ```

## Submitting Changes

Nixopus follows [trunk-based development](https://www.atlassian.com/continuous-delivery/continuous-integration/trunk-based-development).

### Workflow

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make Your Changes**
   - Follow existing code patterns
   - Write tests for new functionality
   - Update documentation if needed
   - Run linters and tests

3. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add your feature"
   # Use conventional commits: feat, fix, docs, test, refactor
   ```

4. **Push and Create Pull Request**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Pull Request Guidelines**
   - Provide a clear description
   - Reference related issues
   - Include tests
   - Update documentation
   - Ensure CI checks pass

### Proposing New Features

1. Check existing issues and PRs
2. Create an issue with the `Feature request` template
3. Include:
   - Feature description
   - Technical implementation details
   - Impact on existing code
   - Use cases

## Code Style and Guidelines

### Backend (Go)

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful names
- Add comments for complex logic
- Handle errors properly
- Write tests for new code

### Frontend (TypeScript/React)

- Use TypeScript for type safety
- Use functional components with hooks
- Follow React best practices
- Use Tailwind CSS for styling
- Ensure accessibility

### General

- Write clear commit messages
- Keep changes focused
- Update related documentation
- Consider performance implications
- Test thoroughly

## Need Help?

If you need assistance:

- Join our [Discord community](https://discord.gg/skdcq39Wpv)
- Open a discussion on [GitHub](https://github.com/raghavyuva/nixopus/discussions)
- Create an issue for bugs or feature requests
- Check existing documentation

Thank you for contributing to Nixopus! Your efforts help make this project better for everyone.
