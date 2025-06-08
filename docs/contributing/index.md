# Contributing to Nixopus

Thank you for your interest in contributing to Nixopus! This guide will help you get started with the development setup and explain the contribution process.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Development Setup](#development-setup)
3. [Running the Application](#running-the-application)
4. [Making Changes](#making-changes)
5. [Submitting a Pull Request](#submitting-a-pull-request)
6. [Proposing New Features](#proposing-new-features)
7. [Extending Documentation](#extending-documentation)
8. [Contributor License Agreement](#contributor-license-agreement)

## Code of Conduct

Before contributing, please review and agree to our [Code of Conduct](/code-of-conduct/index.md). We're committed to maintaining a welcoming and inclusive community.

### Developmentg Setup

If you prefer to set up your development environment manually:

1. Clone the repository:
```bash
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus
```

2. Install Go (version 1.23.6 or newer), and PostgreSQL.

3. Set up PostgreSQL databases:
```bash
createdb postgres -U postgres

createdb nixopus_test -U postgres
```

4. Copy and configure environment variables:
```bash
cp .env.sample .env
```

5. Install project dependencies:
```bash
cd api
go mod download

cd ../view
yarn install
```

## Running the Application
1. Start the API service:
```bash
cd api
air
```

2. Start the view service:
```bash
cd view
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