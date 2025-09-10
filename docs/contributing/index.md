# Contributing to Nixopus

Thank you for your interest in contributing to Nixopus! This guide will help you get started with the development setup and explain the contribution process.

## Table of Contents

- [Contributing to Nixopus](#contributing-to-nixopus)
  - [Table of Contents](#table-of-contents)
  - [Specialized Contribution Guides](#specialized-contribution-guides)
  - [Code of Conduct](#code-of-conduct)
  - [Development Setup](#development-setup)
  - [Running the Application](#running-the-application)
  - [Making Changes](#making-changes)
  - [Submitting a Pull Request](#submitting-a-pull-request)
  - [Proposing New Features](#proposing-new-features)
  - [Extending Documentation](#extending-documentation)
  - [Gratitude](#gratitude)

## Specialized Contribution Guides

We provide detailed guides for specific types of contributions:

- [Getting Started with contribution](README.md) - For general contribution guidelines
- [Backend Development Guide](backend.md) - For Go backend contributions
- [Frontend Development Guide](frontend.md) - For Next.js/React frontend contributions
- [Documentation Guide](documentation.md) - For improving or extending documentation
- [Self-Hosting Guide](self-hosting.md) - For improving installation and self-hosting
- [Docker Guide](docker.md) - For Docker builds and container optimization
- [Development Fixtures Guide](fixtures.md) - For working with development data and fixtures

## Code of Conduct

Before contributing, please review and agree to our [Code of Conduct](/code-of-conduct/index.md). We're committed to maintaining a welcoming and inclusive community.

## Development Setup

If you prefer to set up your development environment manually:

1. Fork the repository: Go to [nixopus GitHub repository](https://github.com/raghavyuva/nixopus). Click on Fork to create your own copy under your GitHub account.

1. Clone the repository:

```bash
git clone git@github.com:your_username/nixopus.git
cd nixopus
```

2. Install Go (version 1.23.6 or newer) and PostgreSQL.

3. Set up PostgreSQL databases:

```bash
createdb nixopus -U postgres
createdb nixopus_test -U postgres
```

4. Copy and configure environment variables (API service):

```bash
cd api
cp .env.sample .env
# Update .env to match your local DB (e.g., DB_NAME=nixopus, USERNAME=postgres, PASSWORD=...)
```

5. Install project dependencies:

```bash
go mod download

cd ../view
yarn install
```

6. Load development fixtures (optional but recommended):

```bash
cd ../api

# Load fixtures without affecting existing data
make fixtures-load

# Or for a clean slate (drops and recreates tables)
make fixtures-recreate

# Get help on fixtures commands
make fixtures-help
```

The fixtures system provides sample data including users, organizations, roles, permissions, and feature flags to help you get started quickly with development.

## Running the Application

1. Start the API service:

```bash
air
```

2. Start the view service:

```bash
cd ../view
yarn dev
```

The view service uses:

- Next.js 15 with App Router
- React 19
- Redux Toolkit for state management
- Tailwind CSS for styling
- Radix UI for accessible components (Shadcn Components)
- TypeScript for type safety

## Making Changes

Nixopus follows [trunk-based-development](https://www.atlassian.com/continuous-delivery/continuous-integration/trunk-based-development) conventions.

1. Create a new branch:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following the project structure:
   - Place new features under `api/internal/features/`
   - Add tests for new functionality
   - Update migrations if needed
   - Follow existing patterns for controllers, services, and storage
   - For frontend changes, follow the Next.js app directory structure

3. Run tests:

```bash
cd api
make test

# View linting
cd view
yarn lint
```

4. Commit your changes with clear messages.

## Submitting a Pull Request

1. Push your branch and create a pull request.

2. Ensure your code:
   - Follows the project structure
   - Includes tests
   - Updates documentation if needed
   - Passes all CI checks

3. Be prepared to address feedback.

## Proposing New Features

1. Check existing issues and pull requests.

2. Create a new issue with the `Feature request` template.

3. Include:
   - Feature description
   - Technical implementation details
   - Impact on existing code

## Extending Documentation

Documentation is located in the `docs/` directory. Follow the existing structure and style when adding new content.

## Gratitude

Thank you for contributing to Nixopus! Your efforts help make this project better for everyone.
