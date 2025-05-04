# [](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.4...v) (2025-05-04)



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
