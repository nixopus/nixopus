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



# [0.1.0-alpha.13](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.12...v0.1.0-alpha.13) (2025-08-24)


### Bug Fixes

* ([#325](https://github.com/raghavyuva/nixopus/issues/325)) typography showing borders for h2 tags and remove domains as the title from the domain page ([bfc7c6e](https://github.com/raghavyuva/nixopus/commit/bfc7c6e1f24e89c990696d1c6bef9dbcdac292cc))
* **build:** mount repo root and set work/cli to include helpers/config.prod.yaml ([#330](https://github.com/raghavyuva/nixopus/issues/330)) ([ed2f63f](https://github.com/raghavyuva/nixopus/commit/ed2f63f4b321464b1fd633ed4351385a0d4845cc))
* **ci:** add wrapper venv for python cli as release version ([#333](https://github.com/raghavyuva/nixopus/issues/333)) ([f14b42a](https://github.com/raghavyuva/nixopus/commit/f14b42aea1e575fcb29781c782100031c1ec1c08))
* **ci:** fix PyInstaller build and run PR builds ([#331](https://github.com/raghavyuva/nixopus/issues/331)) ([130b920](https://github.com/raghavyuva/nixopus/commit/130b92058681148af3e6f56189bee1622f9b9ecc))
* deployment edit page showing duplicate form fields ([5c07f68](https://github.com/raghavyuva/nixopus/commit/5c07f6816493e0ce049722667de0d5f4e72d64ba))
* login with ip address deployments ([da08719](https://github.com/raghavyuva/nixopus/commit/da08719306bb821bbbc5999c6328de0833c1b575))
* resolve vitepress build by modifying copy button to avoid invalid vue attribute quoting ([#345](https://github.com/raghavyuva/nixopus/issues/345)) ([250a967](https://github.com/raghavyuva/nixopus/commit/250a967b3b50ca795a3ed49d75228046a15bb5f9))
* support older glibc versions ([#338](https://github.com/raghavyuva/nixopus/issues/338)) ([ac17507](https://github.com/raghavyuva/nixopus/commit/ac1750753c9be90eb2c74205edc9d8f67e41d1d0))
* syntax issue extra braces removed ([dbe1f7a](https://github.com/raghavyuva/nixopus/commit/dbe1f7abe1100cd525e1522bb703806420eafe3d))
* update release cli workflow ([38075dd](https://github.com/raghavyuva/nixopus/commit/38075dd853281eb3954964a6d25f01fd160c7686))
* websocket connection issues in production ([f7a649a](https://github.com/raghavyuva/nixopus/commit/f7a649af57a44fe1d8b6c88d44b0c71ac77ca7d9))


### Features

* add fetching branches for repository during self hosting  ([#332](https://github.com/raghavyuva/nixopus/issues/332)) ([c480e8b](https://github.com/raghavyuva/nixopus/commit/c480e8be26563ba14babc7d3f712688e22c1ea94))
* add multi stepper form for deployment form ([#327](https://github.com/raghavyuva/nixopus/issues/327)) ([1a161e3](https://github.com/raghavyuva/nixopus/commit/1a161e33189de631b5549f6f807bfbfe45faf426))
* integration of viper configuration management in api ([#311](https://github.com/raghavyuva/nixopus/issues/311)) ([e81d038](https://github.com/raghavyuva/nixopus/commit/e81d038017d79ca0f0cd6699f9b43628afeda8d9))
* merge install scripts, improve theme handling, and enhance container UI ([#328](https://github.com/raghavyuva/nixopus/issues/328)) ([8310aa8](https://github.com/raghavyuva/nixopus/commit/8310aa8fc0fffa0a9410cf6f0b2624d46cfd5b01))
* mobile first file manager component design ([#349](https://github.com/raghavyuva/nixopus/issues/349)) ([d79ea0e](https://github.com/raghavyuva/nixopus/commit/d79ea0ea3510fe593b0a5a8b7abb355928b4c7ff))
* password input field with show/hide toggle button ([#342](https://github.com/raghavyuva/nixopus/issues/342)) ([571f1af](https://github.com/raghavyuva/nixopus/commit/571f1af5c751b9b5f08ef146a6fa037836b72093))



