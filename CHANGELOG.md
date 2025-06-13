# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.8...v) (2025-06-13)


### Bug Fixes

* Domain deployment fails due to unresolved helpers/caddy.json path ([b3bb53c](https://github.com/raghavyuva/nixopus/commit/b3bb53c340f459aadc82ea4117388e43c653cba9))
* installation message to print out ip:port format ([12f0354](https://github.com/raghavyuva/nixopus/commit/12f0354010e0ef4467e961759b14d9d374afde42))
* remove asking for confirmation from user when domain validation fails ([0014e84](https://github.com/raghavyuva/nixopus/commit/0014e846972c3a1f9751e91078688c2e55cb11ce))
* remove interactive admin credential asking through installation wizard ([cfdb159](https://github.com/raghavyuva/nixopus/commit/cfdb1592bfdce041a64b2706a9c101f3c0885925))
* remove string quotes on parameter passing in qemu steps ([73746af](https://github.com/raghavyuva/nixopus/commit/73746af599526c40b202942e05491c256ccf30f8))
* seperate jobs for domain based installation and ip based installation ([b0736ad](https://github.com/raghavyuva/nixopus/commit/b0736ad5fd31c65aeb87d5157a19e282fdcaaeb9))



# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.7...v) (2025-06-11)



# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.7...v) (2025-06-11)



# Changelog
All notable changes to this project will be documented in this file.

# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.6...v) (2025-05-25)
## Fixes
* Fixes Caddy Server Route Duplication #101

## Chore
* [fix(sidebar): remove container feature from allowed resource in sidebar permission checking](https://github.com/raghavyuva/nixopus/pull/106/commits/1cec95d8cf1f5e25e179c1b206e56f648ec02b05) 
* [chore(version): update version for v0.1.0.alpha.6](https://github.com/raghavyuva/nixopus/pull/106/commits/52f2755d9690b8a6ab2498b9d88d3ed302e88dc5)
## Test
* #102 - test(makefile): update test target to run only unit tests from features directory 
* #107 - Test : Add Unit Test for Feature Flags

## Contributors
@raghavyuva @shravan20 

# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.5...v) (2025-05-04)
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

## [v0.1.0-alpha.4]

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