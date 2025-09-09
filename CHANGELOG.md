# [0.1.0-alpha.21](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.20...v0.1.0-alpha.21) (2025-09-08)


### Bug Fixes

* **ci:** format workflow to single-commit, sequential and use dorny/paths-filter ([#374](https://github.com/raghavyuva/nixopus/issues/374)) ([c74e074](https://github.com/raghavyuva/nixopus/commit/c74e07456796fa36331dfa32bba5035cb73712c7))



# [0.1.0-alpha.20](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.19...v0.1.0-alpha.20) (2025-09-07)


### Bug Fixes

* **cli:** add Docker cleanup on force reinstall to ensure fresh stack ([#371](https://github.com/raghavyuva/nixopus/issues/371)) ([1cfe009](https://github.com/raghavyuva/nixopus/commit/1cfe009c4f95193bca73402bd83a2a1c944ca8d0))



# [0.1.0-alpha.19](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.18...v0.1.0-alpha.19) (2025-09-03)


### Features

* container listing with pagination, search, and sort ([#367](https://github.com/raghavyuva/nixopus/issues/367)) ([7400fda](https://github.com/raghavyuva/nixopus/commit/7400fdae767468a44062eb05468349cfa149219c))



# [0.1.0-alpha.18](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.17...v0.1.0-alpha.18) (2025-09-02)


### Bug Fixes

* **ci:** fix format workflow auto-commit on pushes ([#365](https://github.com/raghavyuva/nixopus/issues/365)) ([f74f00a](https://github.com/raghavyuva/nixopus/commit/f74f00a60d78ff4b2632d3442eaec0bccc5219f7))



# [0.1.0-alpha.17](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.16...v0.1.0-alpha.17) (2025-08-29)


### Bug Fixes

* failing test case due to --depth option in clone logic ([f474235](https://github.com/raghavyuva/nixopus/commit/f4742350a123952451cb0db24022a3c5aed92e27))


### Features

* default --config-file to None in command ([62583f7](https://github.com/raghavyuva/nixopus/commit/62583f7d14a157235f3a130fc69c46ddbdefcebf))
* fallback to built-in config when no --config-file is provided ([cd6eafd](https://github.com/raghavyuva/nixopus/commit/cd6eafd56a7bdb10ff1c605677e2fd231c4092fe))
* load built-in config via Config.load_yaml_config() when config_file is None. ([885dbf2](https://github.com/raghavyuva/nixopus/commit/885dbf28b4496f018b29d221d96f7dede970404c))



