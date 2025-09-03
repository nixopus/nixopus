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



# [0.1.0-alpha.16](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.15...v0.1.0-alpha.16) (2025-08-29)


### Bug Fixes

* audit log rendering logic ([#359](https://github.com/raghavyuva/nixopus/issues/359)) ([42cc81e](https://github.com/raghavyuva/nixopus/commit/42cc81e63abb7c01e52b6978b5a498afd8d21144))


### Features

* add table component for containers listing and component seggregation ([#356](https://github.com/raghavyuva/nixopus/issues/356)) ([9674ad0](https://github.com/raghavyuva/nixopus/commit/9674ad0446c3afc1ed723aaa44dc51c833a77bd3))


### Reverts

* auto update feature from dashboard ([#360](https://github.com/raghavyuva/nixopus/issues/360)) ([af22103](https://github.com/raghavyuva/nixopus/commit/af2210364f7ab28a6efc621af73afa41e5d127c7))



# [0.1.0-alpha.15](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.14...v0.1.0-alpha.15) (2025-08-26)


### Bug Fixes

* remove changelog as a seperate action, and uses ssh-key for checkout code ([279d988](https://github.com/raghavyuva/nixopus/commit/279d98857a83be7de1c5bc307abfc1c1a664d8eb))


### Features

* makes use of the ssh push in release action instead of the default behaviour ([3d36258](https://github.com/raghavyuva/nixopus/commit/3d36258c421f30baafd3d6f1c856402122208100))



# [0.1.0-alpha.14](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.13...v0.1.0-alpha.14) (2025-08-26)


### Bug Fixes

* staging compose file to match with the latest cli versioned installation structure ([ddcf648](https://github.com/raghavyuva/nixopus/commit/ddcf648ab07085179a023e44e0024e3838ac4423))
* update version to v0.1.0-alpha.13 ([3cd82f1](https://github.com/raghavyuva/nixopus/commit/3cd82f181bdc9fdb9e9454556cd187392bc2acd4))


### Features

* upgrade Nixopus install script with detailed usage, extended CLI options ([#351](https://github.com/raghavyuva/nixopus/issues/351)) ([356eb25](https://github.com/raghavyuva/nixopus/commit/356eb2531ce0b8176128e44c5dadfd9bc08a344f))



