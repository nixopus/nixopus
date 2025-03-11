# Contributing to Nixopus

Thank you for your interest in contributing to Nixopus! This guide will help you get started with the development setup and explain the contribution process.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Development Setup](#development-setup)
   - [Using Dev Container](#using-dev-container)
   - [Manual Setup](#manual-setup)
3. [Running the Application](#running-the-application)
4. [Making Changes](#making-changes)
5. [Submitting a Pull Request](#submitting-a-pull-request)
6. [Proposing New Features](#proposing-new-features)
7. [Extending Documentation](#extending-documentation)
8. [Contributor License Agreement](#contributor-license-agreement)

## Code of Conduct

Before contributing, please review and agree to our [Code of Conduct](/code-of-conduct/index.md). We're committed to maintaining a welcoming and inclusive community.

## Development Setup
### Using Dev Container

For a quick and easy setup, you can use the provided Dev Container configuration.
This method requires VS Code and Docker to be installed on your system.

1. Click [here](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/nixopus/nixopus) to clone the repository and open it in a Dev Container.
2. VS Code will automatically install the Dev Containers extension if needed, clone the source code into a container volume, and spin up a dev container for use.

The Dev Container is configured with all necessary dependencies and tools.
Once it's ready, you can start developing right away.

### Manual Setup

If you prefer to set up your development environment manually, follow these steps:

1. Clone the repository and switch to nixopus directory:

    `git clone https://github.com/nixopus/nixopus.git` 

    `cd nixopus`

2. Install [Node.js](https://nodejs.org/en/download/package-manager) (version 18.10 or newer) and [yarn](https://classic.yarnpkg.com/lang/en/docs/install/) (version 4.5.0 or newer) .

3. Install Rust and Cargo (latest stable version) click [here](https://doc.rust-lang.org/cargo/getting-started/installation.html).

4. Install PostgreSQL and libpq-dev:

    `sudo apt-get update && sudo apt-get install -y postgresql libpq-dev`

5. Install Diesel CLI:

    `cargo install diesel_cli --no-default-features --features postgres`

6. Install cargo-watch:

    `cargo install cargo-watch`

7. Set up the required environment variables (you may want to add these to your `.env` file):
```
export POSTGRES_DB=mydb
export POSTGRES_USER=myuser
export POSTGRES_PASSWORD=mypassword
export DATABASE_URL=postgres://myuser:mypassword@domain:port/mydb
export CORE_PORT=8080
export RUST_LOG=info
export NEXT_PUBLIC_API_URL=http://localhost:8080
export JWT_SECRET=change_this_secret
export SERVER_ADDR=0.0.0.0
export SERVER_PORT=8080
export FILE_STORAGE_PATH=/
export HASH_SECRET=somehashscecrethere
export RUST_BACKTRACE=1
export GITHUB_APP_PRIVATE_KEY=""
export GITHUB_APP_ID=238234
export ZITADEL_APPLICATION='{"type":"application","keyId":"","key":","appId":"","clientId":""}'
```

8. Install project dependencies:

`cd app && yarn install`

## Running the Application

The Nixopus project consists of multiple components. Here's how to run each part:

1. Start the PostgreSQL database (if not already running).

2. Run the core Rust application:

`cd core && cargo watch -x run`

3. Run the consumer application:

`cd consumer && cargo watch -x run`

4. Run the terminal application:

`cd terminal && cargo watch -x run`

5. Run the Next.js application:

`cd app && yarn run dev`

## Making Changes
Nixopus follows [trunk-based-development](https://www.atlassian.com/continuous-delivery/continuous-integration/trunk-based-development) conventions. It's recommended to follow the same to make all the commit messages and feature/fixes more clear. 

1. Create a new branch for your changes. For example, if you're working on a feature, name it `feature/your-feature-name`.

`git checkout -b feature/your-feature-name`

2. Make your changes and commit them with clear, concise commit messages.

3. Push your changes to your fork:

`git push origin feature/your-feature-name`

## Submitting a Pull Request

1. Go to the Nixopus repository on GitHub and click the `New pull request` button.

2. Select your branch and provide a clear title and description for your pull request.

3. Ensure that your code follows the project's coding standards and best practices.

4. Include tests for your changes if applicable.

5. Update the documentation if your changes affect the user-facing functionality.

6. Be prepared to respond to feedback and make necessary adjustments.

## Proposing New Features

If you have an idea for a new feature:

1. Check the existing issues and pull requests to see if it has already been proposed.

2. If not, create a new issue with the `Feature request` template.

3. Clearly describe the feature, its benefits, and potential implementation details.

4. Engage in discussion with maintainers and other contributors about the feature.

## Extending Documentation

To contribute to the documentation:

1. The documentation repository is located at [https://github.com/nixopus/nixopus](https://github.com/nixopus/nixopus).

2. Follow the same process as code contributions for submitting documentation changes.

3. Ensure your writing is clear, concise, and follows the existing documentation style.

## Contributor License Agreement

Before your contributions can be merged, you need to sign our [Contributor License Agreement](CONTRIBUTOR_LICENSE_AGREEMENT.md). This is a simple process:

1. When you open a pull request, an automated bot will comment with instructions.

2. Follow the link provided by the bot to sign the CLA electronically.

3. Once signed, the bot will update the pull request, allowing it to be merged.

Thank you for contributing to Nixopus! Your efforts help make this project better for everyone.