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



# [0.1.0-alpha.12](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.11...v0.1.0-alpha.12) (2025-08-09)


### Bug Fixes

* branch-rule on release cli ([#318](https://github.com/raghavyuva/nixopus/issues/318)) ([2ed3d17](https://github.com/raghavyuva/nixopus/commit/2ed3d172a498bde12f20f6401ae9ac84b02cdaf2))
* change release cli naming issue in workflow path ([353777e](https://github.com/raghavyuva/nixopus/commit/353777ed3cf56b0ae9ba84553e8e29e3c23116a8))
* update release branch to trigger on master push ([adfdeba](https://github.com/raghavyuva/nixopus/commit/adfdeba581d3a21d0c3f8155c953a7e583958016))



# [0.1.0-alpha.11](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.9...v0.1.0-alpha.11) (2025-07-13)


### Bug Fixes

* add permissions for dashboard, and terminal features, wrapped under rbac guard' ([69c37c1](https://github.com/raghavyuva/nixopus/commit/69c37c138af965292041fe84402cc75d62ee51e1))
* changelog push to pull request ([cc35929](https://github.com/raghavyuva/nixopus/commit/cc3592943fe609b0a0eb9a2505c46a8746a028e7))
* command chaining in contributing docs ([7799d4d](https://github.com/raghavyuva/nixopus/commit/7799d4d02569b36346a051fc90f3fb3e0cc78020))
* **docs:** fix incorrect method display, correct extraction logic, and update VitePress sidebar link ([#220](https://github.com/raghavyuva/nixopus/issues/220)) ([2c5d490](https://github.com/raghavyuva/nixopus/commit/2c5d490217a8bc50f182dc954964bb2368e6b693))
* **docs:** preview open API docs in documentation ([#224](https://github.com/raghavyuva/nixopus/issues/224)) ([24d196c](https://github.com/raghavyuva/nixopus/commit/24d196ca706278d47101114918da90ec3949ea23))
* notification feature broken due to rbac guard implementation ([be1c1f8](https://github.com/raghavyuva/nixopus/commit/be1c1f8480b5fb09c999567e3aaf64a86bffeba3))
* remove fallback to access denied component when something is not passed to rbac related guard as props ([c1b6ad4](https://github.com/raghavyuva/nixopus/commit/c1b6ad426ccf7608583765c117848b23d349faa0))
* remove macos related inconsistency in dev env setup action file ([fb812af](https://github.com/raghavyuva/nixopus/commit/fb812affbf12f02e879b5b02b8681428147e1df0))
* update changelog workflow to include only master branch push ([b0e38dc](https://github.com/raghavyuva/nixopus/commit/b0e38dccc18de00301abcc6b87400deeee0731c6))
* update release workfflow not to push rather create a pr with changes ([a554c87](https://github.com/raghavyuva/nixopus/commit/a554c87f1fa143a049ccf48c3327d1cd357fa975))
* update workflows to be more specific on the events thus by making better use of actions' ([a1a144b](https://github.com/raghavyuva/nixopus/commit/a1a144b71d363f818868474e89eba3673a866a33))
* uses permission guard to have more type safe declarations ([813924c](https://github.com/raghavyuva/nixopus/commit/813924c22ea17988e6c1fd6fb91e2ed94ea3fce4))


### Features

* create rbac guard and util components for different combination of permission checks ([df5873d](https://github.com/raghavyuva/nixopus/commit/df5873d0ed47b4696638972cc301e65bca8b4798))
* enable auto update of version.txt on release ([9a554f7](https://github.com/raghavyuva/nixopus/commit/9a554f741b935f267b342aadeb8ccb8e45727f78))


### Reverts

* Revert "fix(docs): fix incorrect method display, correct extraction logic, an…" (#223) ([4249422](https://github.com/raghavyuva/nixopus/commit/4249422b64a0e22945f50b5e64e1e0471c5ffe4f)), closes [#223](https://github.com/raghavyuva/nixopus/issues/223)
* temporary changes to install branch ([e4b6759](https://github.com/raghavyuva/nixopus/commit/e4b6759b1b58a5993c33410e839a71be319054cf))



