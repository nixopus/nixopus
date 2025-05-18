# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.5...v) (2025-05-18)

# 📦 Changelog Highlights

## ✨ Features

- **Avatar Upload**
  - Users can now upload a profile avatar to personalize their account.

- **DevContainer Restructure**
  - Improved Dockerfile structure.
  - `air` tool added for hot reloads.
  - Workspace config now uses `localWorkspaceFolder` for better volume mapping.


## 🧪 Tests

- **Integration Test Suite**
  - Set up integration tests using the `go-hit` framework.
  - Covered endpoints:
    - `POST /register`
    - `POST /login`
    - `POST /refresh-token`
    - `POST /reset-password`
    - `GET /user-details`
    - 2FA, email verification, and more.
  - Ensures end-to-end correctness of the authentication flow and user management.


## 🐛 Fixes

- **Docker-Compose & ENV Handling**
  - `.env` file now correctly referenced relative to the source directory.
  - Fixed repeated issues with environment path mismatches.

- **TLS and Port Conflicts**
  - Resolved Docker TLS configuration bugs during install.
  - Updated installer to avoid hardcoded Docker context ports (`2376`).
  - Ensured services start reliably with corrected Docker context.

- **Self-Host Port Mapping**
  - Corrected Caddy port mapping issues.
  - Prevented accidental overwrites of Caddy configs on staging reloads.


## 🔧 Chores

- **Changelog Automation**
  - Automated changelog updates integrated via GitHub Actions.

- **License**
  - Updated license to FSL and added documentation accordingly.

_Contributors: @raghavyuva, @kishore1919, @github-actions_


# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.4...v) (2025-05-04)



# Changelog

All notable changes to this project will be documented in this file.

## [v0.1.0-alpha.4] - Unreleased

### Features
- Users can enable or disable features in Nixopus as needed, giving them full control based on their individual preferences and use cases (#66)
- Added caching layer for API endpoints to improve performance by caching frequently accessed middleware checks (#69)
- Added container management interface with detailed view and actions (restart, stop, remove) (#72)

### Fixes
- Skip automatic update checks and disable update functionality in development environment (#71)
- Automatic port mapping for docker configurations (#73)

### Tests
- Added unit tests for audit, auth, domain features service and storage layers (#68)

### CI
- Spin up Test CI during feat/develop or master branches as the target and for PR's
- Fixed auto commit not working for formatting nixopus api and nixopus view

### Chore
- Added common.loading translations to i18n files
- Prevent domain belongs to server checking logic from Dev Environment
- Removed log viewer components description

[Unreleased]: https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.3...HEAD
[v0.1.0-alpha.4]: https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.3...v0.1.0-alpha.4
