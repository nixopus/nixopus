# [0.1.0-alpha.70](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.68...v0.1.0-alpha.70) (2025-12-01)


### Features

* bump to alpha-v69 to mark release ([#592](https://github.com/raghavyuva/nixopus/issues/592)) ([a521d60](https://github.com/raghavyuva/nixopus/commit/a521d602e1e756c071128c7731a4e8f65bcc2a13))
* include version.txt to bundler ([a986b9a](https://github.com/raghavyuva/nixopus/commit/a986b9a72132c524d7e8d561845eb26989547371))
* read version from installed pkg with fallback to bundler/src ([f79d808](https://github.com/raghavyuva/nixopus/commit/f79d808fff46db0cffb5a7260c433dad60298fb2))
* sudo requirement non root user & improve installation error handling ([#589](https://github.com/raghavyuva/nixopus/issues/589)) ([d56f902](https://github.com/raghavyuva/nixopus/commit/d56f902bd532ab9c7c271ac3f2224f5d077a466a))



# [0.1.0-alpha.68](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.67...v0.1.0-alpha.68) (2025-11-24)


### Bug Fixes

* feature disabled error on signup ([#587](https://github.com/raghavyuva/nixopus/issues/587)) ([8af20ab](https://github.com/raghavyuva/nixopus/commit/8af20abf6c02706e6e726e63da4df7a1399645a3))



# [0.1.0-alpha.67](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.66...v0.1.0-alpha.67) (2025-11-21)


### Features

* compose as extensions ([#555](https://github.com/raghavyuva/nixopus/issues/555)) ([741aa6a](https://github.com/raghavyuva/nixopus/commit/741aa6ab30520f46cc796c6510ea9c2551c4fd8e))



# [0.1.0-alpha.66](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.65...v0.1.0-alpha.66) (2025-11-21)


### Bug Fixes

* allow custom ports on install setup optionally ([#580](https://github.com/raghavyuva/nixopus/issues/580)) ([972c7ac](https://github.com/raghavyuva/nixopus/commit/972c7ac4ea2aedd7810954772c4d16d7226182d6))



# [0.1.0-alpha.65](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.64...v0.1.0-alpha.65) (2025-11-09)


### Bug Fixes

* remove linux/arm/v7 support since no native support from postcss and Nextjs sharp ([#576](https://github.com/raghavyuva/nixopus/issues/576)) ([dc84f0e](https://github.com/raghavyuva/nixopus/commit/dc84f0ef1c1a2568885fdd02e144b91ebe19d8a1))



# [0.1.0-alpha.64](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.63...v0.1.0-alpha.64) (2025-11-08)


### Bug Fixes

* update docker compose files to use internal ports for supertokens postgres connection ([9a23b2c](https://github.com/raghavyuva/nixopus/commit/9a23b2c4552b4f6905dd6d501ed0a40cddf362c0))



# [0.1.0-alpha.63](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.62...v0.1.0-alpha.63) (2025-11-08)


### Bug Fixes

* malformed supertokens connection uri during ip based installations ([b6b1ad5](https://github.com/raghavyuva/nixopus/commit/b6b1ad5ec11b3eda9caac692c0cad3cba59adbc9))



# [0.1.0-alpha.62](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.61...v0.1.0-alpha.62) (2025-11-08)


### Bug Fixes

* always binds predictable ports inside the container, and uses dynamic ports for the host ([#569](https://github.com/raghavyuva/nixopus/issues/569)) ([e5e637e](https://github.com/raghavyuva/nixopus/commit/e5e637eabddfa02ed03b32de338004ed5efcfaa8))


### Features

* add support for linux/amd64 linux/arm54 linux/arm/v7 ([#570](https://github.com/raghavyuva/nixopus/issues/570)) ([59938ad](https://github.com/raghavyuva/nixopus/commit/59938ad28932854c9b0d2388b37ceeb28a8a1ab9))



# [0.1.0-alpha.61](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.60...v0.1.0-alpha.61) (2025-11-08)


### Features

* add support for custom ports during nixopus install ([#567](https://github.com/raghavyuva/nixopus/issues/567)) ([01c4b1d](https://github.com/raghavyuva/nixopus/commit/01c4b1d8a116fde1cf992cef4face3728e55e039))



# [0.1.0-alpha.60](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.59...v0.1.0-alpha.60) (2025-11-05)


### Bug Fixes

* nixopus uninstall hangs or silently asks for confirmation from user which is not identical ([#560](https://github.com/raghavyuva/nixopus/issues/560)) ([5a9c7f8](https://github.com/raghavyuva/nixopus/commit/5a9c7f8cb7ec347a907ba51de4a9e6bc59f707ec))


### Features

* add support for custom config file during nixopus installation ([#561](https://github.com/raghavyuva/nixopus/issues/561)) ([0b34f84](https://github.com/raghavyuva/nixopus/commit/0b34f84e02345f2bb583be587c7f4676b72c6523))



# [0.1.0-alpha.59](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.58...v0.1.0-alpha.59) (2025-11-04)


### Features

* add support for custom ip address deployments ([#554](https://github.com/raghavyuva/nixopus/issues/554)) ([d1fb0b4](https://github.com/raghavyuva/nixopus/commit/d1fb0b42fdfad3709180c0ffb1d4725d1e5c8e7b))



# [0.1.0-alpha.58](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.57...v0.1.0-alpha.58) (2025-11-02)


### Features

* port linux server images to extension templates ([#556](https://github.com/raghavyuva/nixopus/issues/556)) ([c1bcb7e](https://github.com/raghavyuva/nixopus/commit/c1bcb7e79edb4819abe8ccef0704c2c841ca6671))



# [0.1.0-alpha.57](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.56...v0.1.0-alpha.57) (2025-11-02)


### Features

* improve dashboard monitoring & reset cache feature flags ([#557](https://github.com/raghavyuva/nixopus/issues/557)) ([fdf26bb](https://github.com/raghavyuva/nixopus/commit/fdf26bb9d84d5499912a8ce5e88a07e9e95d8655))



# [0.1.0-alpha.56](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.55...v0.1.0-alpha.56) (2025-10-30)


### Features

* add extension templates for gotify, n8n, netdata, qdrant, and more ([#545](https://github.com/raghavyuva/nixopus/issues/545)) ([ecb332c](https://github.com/raghavyuva/nixopus/commit/ecb332c51b0f99beff625856f2ff7a7ec4e9d33c))



# [0.1.0-alpha.55](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.54...v0.1.0-alpha.55) (2025-10-29)


### Bug Fixes

* make domain validation less restrictive for extension deployments ([#543](https://github.com/raghavyuva/nixopus/issues/543)) ([72cc971](https://github.com/raghavyuva/nixopus/commit/72cc971f0f8df0897c60b31379090248ec771f74))



# [0.1.0-alpha.54](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.53...v0.1.0-alpha.54) (2025-10-28)


### Features

* configurable dashboard widgets with topbar  ([#541](https://github.com/raghavyuva/nixopus/issues/541)) ([b150d69](https://github.com/raghavyuva/nixopus/commit/b150d6937db92fa288b40bffe54f4579a95f252a))



# [0.1.0-alpha.53](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.52...v0.1.0-alpha.53) (2025-10-27)


### Features

* dashboard with draggable layout, charts, and extended system metrics ([#536](https://github.com/raghavyuva/nixopus/issues/536)) ([e13c24a](https://github.com/raghavyuva/nixopus/commit/e13c24aeefeb67fb64da1809bc23695f3076bf46))



# [0.1.0-alpha.52](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.51...v0.1.0-alpha.52) (2025-10-25)


### Bug Fixes

* disable just in time compilation (JIT) of postgres ([#539](https://github.com/raghavyuva/nixopus/issues/539)) ([b2c35bd](https://github.com/raghavyuva/nixopus/commit/b2c35bd29349f565aa66617a164d687faf060778))



# [0.1.0-alpha.51](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.50...v0.1.0-alpha.51) (2025-10-25)


### Bug Fixes

* menu for closed sidebar items on hover ([#526](https://github.com/raghavyuva/nixopus/issues/526)) ([ca423ed](https://github.com/raghavyuva/nixopus/commit/ca423ed2e3a53c5a1e96048914316399274afcf0))



# [0.1.0-alpha.50](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.49...v0.1.0-alpha.50) (2025-10-25)


### Features

* update command to not reference .env ([af13242](https://github.com/raghavyuva/nixopus/commit/af13242c25e2cc8b02e965e8b3645df84e372c9b))



# [0.1.0-alpha.49](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.48...v0.1.0-alpha.49) (2025-10-22)


### Bug Fixes

* borders not visible in light themes ([#525](https://github.com/raghavyuva/nixopus/issues/525)) ([8756ff7](https://github.com/raghavyuva/nixopus/commit/8756ff7e1b81672df98670e2f592ca73b4adab98))



# [0.1.0-alpha.48](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.47...v0.1.0-alpha.48) (2025-10-22)


### Features

* **cli:** bump cli v0.1.15 to v0.1.16 ([#529](https://github.com/raghavyuva/nixopus/issues/529)) ([0db449b](https://github.com/raghavyuva/nixopus/commit/0db449b5fa5b8235fffe12b09b3dbdcfaecccf9a))



# [0.1.0-alpha.47](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.46...v0.1.0-alpha.47) (2025-10-22)


### Features

* **cli:** live reloading dockerized dev setup ([#522](https://github.com/raghavyuva/nixopus/issues/522)) ([a05a0d6](https://github.com/raghavyuva/nixopus/commit/a05a0d658ba42284404e0e72d930630abd5a74d1))



# [0.1.0-alpha.46](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.45...v0.1.0-alpha.46) (2025-10-22)


### Features

* **terminal:** support clipboard for terminal input/output ([#515](https://github.com/raghavyuva/nixopus/issues/515)) ([8ad6a1c](https://github.com/raghavyuva/nixopus/commit/8ad6a1c08c64eebfbfad5c83bd506ce8bc3fd508))



# [0.1.0-alpha.45](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.44...v0.1.0-alpha.45) (2025-10-22)


### Features

* nixopus update ([#401](https://github.com/raghavyuva/nixopus/issues/401)) ([3913d60](https://github.com/raghavyuva/nixopus/commit/3913d60a1566f86a71407103417ce5fabd35a086))



# [0.1.0-alpha.44](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.43...v0.1.0-alpha.44) (2025-10-18)


### Features

* add custom domain support for templates ([956e889](https://github.com/raghavyuva/nixopus/commit/956e8892db6736b5c70b4099497612a828d2369e))



# [0.1.0-alpha.43](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.42...v0.1.0-alpha.43) (2025-10-18)


### Features

* add proxy support for extensions ([88ce1bc](https://github.com/raghavyuva/nixopus/commit/88ce1bcaded066ad8f906f03731e1f3ea925f908))



# [0.1.0-alpha.42](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.41...v0.1.0-alpha.42) (2025-10-17)


### Features

* setup development environment with cli installer ([#508](https://github.com/raghavyuva/nixopus/issues/508)) ([a3647c6](https://github.com/raghavyuva/nixopus/commit/a3647c6f47bc2a75b1367a146a13ed143daedfa6))



# [0.1.0-alpha.41](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.40...v0.1.0-alpha.41) (2025-10-16)


### Bug Fixes

* security scan to have TRIVY_DISABLE_VEX_NOTICE ([43546b8](https://github.com/raghavyuva/nixopus/commit/43546b88b507aa776e2260d24a1ac68594630fe0))



# [0.1.0-alpha.40](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.39...v0.1.0-alpha.40) (2025-10-15)


### Bug Fixes

* db getting wiped due to hosts permission issue, switches back to named docker maintained volumes ([#507](https://github.com/raghavyuva/nixopus/issues/507)) ([f8fd796](https://github.com/raghavyuva/nixopus/commit/f8fd7964da4c0a69b67c2696fd25694433238718))
* update dockerfile to copy extensions templates folder in production ([5492582](https://github.com/raghavyuva/nixopus/commit/5492582a8c249cd1a169cb8c0ac9615cb67c2984))



# [0.1.0-alpha.39](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.38...v0.1.0-alpha.39) (2025-10-15)


### Bug Fixes

* supertokens connection URI handling for ip addr and domains ([#503](https://github.com/raghavyuva/nixopus/issues/503)) ([9d62c8d](https://github.com/raghavyuva/nixopus/commit/9d62c8d1d3317a0fef008945e09982b0429ad487))



# [0.1.0-alpha.38](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.37...v0.1.0-alpha.38) (2025-10-15)


### Bug Fixes

* extension step execution ([3b03319](https://github.com/raghavyuva/nixopus/commit/3b033195f4c856a3c050e8e5b7c8c79c329f1a5e))
* overflow of descriptions with readmore option ([dab77db](https://github.com/raghavyuva/nixopus/commit/dab77db2868edd1eb49d81489e01579c5c61e2c2))
* rbac permissions according to supertokens changes ([4c3db53](https://github.com/raghavyuva/nixopus/commit/4c3db539e37f497c96943c1b79e4d8b024cacca4))
* search and sorting in extensions ([9558b28](https://github.com/raghavyuva/nixopus/commit/9558b28fd8715307a2358784c21cd33fb0256aef))
* wrap extension page and sidebar in feature flag and rbac guards ([340ff69](https://github.com/raghavyuva/nixopus/commit/340ff696dec012a925603c92135e2530a9211da7))


### Features

* add button for install / run in extension detail page ([e88d011](https://github.com/raghavyuva/nixopus/commit/e88d011988c82d79ba60290889a103b6cd605188))
* add deploy templates ([1bd2fa0](https://github.com/raghavyuva/nixopus/commit/1bd2fa0c034bcdfb8a42db21543c7f2d9496ca01))
* add migrations for extensions permissions, auditing, feature flags ([5d302ac](https://github.com/raghavyuva/nixopus/commit/5d302ac41128fedf21dbb9f09b5059bda41768a5))
* add routes for listing extensions ([9a5e87a](https://github.com/raghavyuva/nixopus/commit/9a5e87a51b6b42216f58626a4b7e919fd7639cb4))
* allow pagination search, sorting and integrate with view ([882c741](https://github.com/raghavyuva/nixopus/commit/882c74159ead7c01c6f4ee379fcc0ee3e264802a))
* define migration for extensions ([a0c64aa](https://github.com/raghavyuva/nixopus/commit/a0c64aab8bdbba5107f4d3b60fe6ab6c460013b0))
* display of status colors based on extension running ([9e490dc](https://github.com/raghavyuva/nixopus/commit/9e490dcce6598c72ea4364b1eb7af4d55279699c))
* enable extension execution with run and cancel apis ([#455](https://github.com/raghavyuva/nixopus/issues/455)) ([9572671](https://github.com/raghavyuva/nixopus/commit/957267176c9804d8066ab6fe03f5d0563f467baa))
* extension category as badges ([0e5d58d](https://github.com/raghavyuva/nixopus/commit/0e5d58ddf4da2ab678c6aa68786f35b7c6489a5a))
* extension details ([#470](https://github.com/raghavyuva/nixopus/issues/470)) ([7fcee25](https://github.com/raghavyuva/nixopus/commit/7fcee25eaa17679a5828eb073fc438dafbd2c296))
* extension discovery and saving to database on api init ([12661f0](https://github.com/raghavyuva/nixopus/commit/12661f0a6e97ea194688035c087597faa4b6cb91))
* extension forking ([#464](https://github.com/raghavyuva/nixopus/issues/464)) ([76238a7](https://github.com/raghavyuva/nixopus/commit/76238a701835a5df4b8c810f4f21c8103625be70))
* extensions ui design with dummy data ([798fef9](https://github.com/raghavyuva/nixopus/commit/798fef94b685e4280a872e56b75b5d85927d9e38))
* log extension execution ([fac665f](https://github.com/raghavyuva/nixopus/commit/fac665f50fdcba7f2335ca573ca58dc72fd72f6e))
* refactor extension executor ([9c89d61](https://github.com/raghavyuva/nixopus/commit/9c89d6103e1a9df938d0414846987163a9df537c))
* rename extension permission migration files ([52d3331](https://github.com/raghavyuva/nixopus/commit/52d3331836ffa60dd3890e8c25cedf688d584942))



# [0.1.0-alpha.37](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.36...v0.1.0-alpha.37) (2025-10-12)


### Bug Fixes

* feature flag ui and feature flag writes missing RBAC permissions ([#493](https://github.com/raghavyuva/nixopus/issues/493)) ([2e1c857](https://github.com/raghavyuva/nixopus/commit/2e1c857231d4587e736a157974585c670e0e09a4))



# [0.1.0-alpha.36](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.35...v0.1.0-alpha.36) (2025-10-11)


### Bug Fixes

* supertokens URI for ip vs domain ([#489](https://github.com/raghavyuva/nixopus/issues/489)) ([aaddb3c](https://github.com/raghavyuva/nixopus/commit/aaddb3c06192ebc1df203690f754ec1b26280134))



# [0.1.0-alpha.35](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.34...v0.1.0-alpha.35) (2025-10-11)


### Bug Fixes

* **cli:** force HTTP protocol for SuperTokens connection URI ([#487](https://github.com/raghavyuva/nixopus/issues/487)) ([eb2c0dd](https://github.com/raghavyuva/nixopus/commit/eb2c0ddb9e3df80e606933082a92a66ef65c24cd))



# [0.1.0-alpha.34](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.33...v0.1.0-alpha.34) (2025-10-11)


### Bug Fixes

* supertokens api url in appinfo.ts ([#486](https://github.com/raghavyuva/nixopus/issues/486)) ([901df3f](https://github.com/raghavyuva/nixopus/commit/901df3f92859f1dbc9bc644041da7e5d7436979e))



# [0.1.0-alpha.33](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.32...v0.1.0-alpha.33) (2025-10-11)


### Bug Fixes

* env config for psql setup with supertokens ([#483](https://github.com/raghavyuva/nixopus/issues/483)) ([5e8db05](https://github.com/raghavyuva/nixopus/commit/5e8db05b8c4b20b729952a79f6c1edbff32bf6db))



# [0.1.0-alpha.32](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.31...v0.1.0-alpha.32) (2025-10-10)


### Features

* integrate SuperTokens authentication system ([#440](https://github.com/raghavyuva/nixopus/issues/440)) ([3e2b678](https://github.com/raghavyuva/nixopus/commit/3e2b6780b3830462fbc5490771ea037d9a1f9c96))



# [0.1.0-alpha.31](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.30...v0.1.0-alpha.31) (2025-10-08)


### Bug Fixes

* **ci:** discord notification on webhooks ([a8261f9](https://github.com/raghavyuva/nixopus/commit/a8261f914aea1481021bd60656036439b88fc8a2))
* **i18n:** update terms phrasing for clarity in English locale ([#460](https://github.com/raghavyuva/nixopus/issues/460)) ([0b96b29](https://github.com/raghavyuva/nixopus/commit/0b96b29a459b8db0414d01c3e720e525e536d6c4))
* **terminal:** custom key event handler for Ctrl + J ([#459](https://github.com/raghavyuva/nixopus/issues/459)) ([291bec7](https://github.com/raghavyuva/nixopus/commit/291bec7e44a577a539286fefa3df8c73fdc997c9))


### Features

* automated discord notifications for new releases ([#439](https://github.com/raghavyuva/nixopus/issues/439)) ([180f299](https://github.com/raghavyuva/nixopus/commit/180f299fe935d04242bf39ba2843fb925cc91910))
* **i18n:** add support to malayalam ([#420](https://github.com/raghavyuva/nixopus/issues/420)) ([0a919b2](https://github.com/raghavyuva/nixopus/commit/0a919b2d4e312f838822891ce0ab88f8a1d817e0))



# [0.1.0-alpha.30](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.29...v0.1.0-alpha.30) (2025-09-19)


### Bug Fixes

* prevent PasswordInputField type override ([#417](https://github.com/raghavyuva/nixopus/issues/417)) ([ad621d9](https://github.com/raghavyuva/nixopus/commit/ad621d9284340495bc5125abe7ea6106d8f38029))
* reassign port in caddy when container gets new port ([60c8f6f](https://github.com/raghavyuva/nixopus/commit/60c8f6fcfca08e68a73119712ff5d218b7dcc41c))


### Features

* cluster based deployment, rollback, restart across services, and more methods wrapper for future integrations for multi server management ([27a8f7a](https://github.com/raghavyuva/nixopus/commit/27a8f7a14125074fbb7dd54c01d08cf3f8d260e0))



# [0.1.0-alpha.29](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.28...v0.1.0-alpha.29) (2025-09-15)


### Bug Fixes

* **ui:** Open Channels tab by default in Notification Settings ([#398](https://github.com/raghavyuva/nixopus/issues/398)) ([3689cd3](https://github.com/raghavyuva/nixopus/commit/3689cd3ca91a3525e1dddc88fd74338c195d5477))



# [0.1.0-alpha.28](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.27...v0.1.0-alpha.28) (2025-09-11)


### Features

* exec commands on container ([#399](https://github.com/raghavyuva/nixopus/issues/399)) ([3cb776d](https://github.com/raghavyuva/nixopus/commit/3cb776daa77dc8df7817df19ff1e0e8a2727788d))
* TaskQ tuning for complete deployment lifecycle ([#393](https://github.com/raghavyuva/nixopus/issues/393)) ([49fe8e1](https://github.com/raghavyuva/nixopus/commit/49fe8e191e9e0a8c0a4480e5d36a21cb3dbf8d11))



# [0.1.0-alpha.27](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.26...v0.1.0-alpha.27) (2025-09-10)


### Features

* install with different branches / forked repositories ([#391](https://github.com/raghavyuva/nixopus/issues/391)) ([8a15b5c](https://github.com/raghavyuva/nixopus/commit/8a15b5c3399a6e2854915e3be6e4ab718fe4a575))



# [0.1.0-alpha.26](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.25...v0.1.0-alpha.26) (2025-09-10)


### Features

* add redis service in docker compose for taskQ ([#386](https://github.com/raghavyuva/nixopus/issues/386)) ([f0a55f1](https://github.com/raghavyuva/nixopus/commit/f0a55f1bef0f1b3119a7ccbabe37030bb4b3ffe6))



# [0.1.0-alpha.25](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.24...v0.1.0-alpha.25) (2025-09-10)


### Bug Fixes

* go sum and go mod conflicts ([5efb26d](https://github.com/raghavyuva/nixopus/commit/5efb26db2e24cdd52d0979041bc300c34096ad81))



# [0.1.0-alpha.24](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.23...v0.1.0-alpha.24) (2025-09-10)


### Features

* migrate async tasks to queue setup via taskq ([#385](https://github.com/raghavyuva/nixopus/issues/385)) ([528c6dc](https://github.com/raghavyuva/nixopus/commit/528c6dcee3554bb6ce38a40896cd6d03a4574ff4))



# [0.1.0-alpha.23](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.22...v0.1.0-alpha.23) (2025-09-10)


### Features

* **notification:** handle smtpConfigs not found ([#384](https://github.com/raghavyuva/nixopus/issues/384)) ([3a3a2a8](https://github.com/raghavyuva/nixopus/commit/3a3a2a897623e729f2680c60425403045d12d125))



# [0.1.0-alpha.22](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.21...v0.1.0-alpha.22) (2025-09-09)


### Bug Fixes

* replacing the input password field with reusable component ([#380](https://github.com/raghavyuva/nixopus/issues/380)) ([2800515](https://github.com/raghavyuva/nixopus/commit/28005150f4486fd5e98493cc5db26f153ae80bc0))



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



# [0.1.0-alpha.9](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.8...v0.1.0-alpha.9) (2025-06-25)


### Bug Fixes

* ([#179](https://github.com/raghavyuva/nixopus/issues/179)) update installation script URL from 'main' to 'master' branch  ([c45e9e7](https://github.com/raghavyuva/nixopus/commit/c45e9e7841e0f5505a230c0a790baf5ef5ce7ee5))
* add go version check ([9c36b59](https://github.com/raghavyuva/nixopus/commit/9c36b59b2e9a30901fa4910387a840135069f086))
* add missing install air hot reload function to main ([00e2554](https://github.com/raghavyuva/nixopus/commit/00e255450d5cfaa5c55f98af82f7c1e42f1b8fad))
* address review comments ([3359adf](https://github.com/raghavyuva/nixopus/commit/3359adf4ae943d0ac7e4121b0d874c347000a7ea))
* Domain deployment fails due to unresolved helpers/caddy.json path ([b3bb53c](https://github.com/raghavyuva/nixopus/commit/b3bb53c340f459aadc82ea4117388e43c653cba9))
* env sample loading issues ([623f8f1](https://github.com/raghavyuva/nixopus/commit/623f8f1980cebcd386f3ce7c6469c51c84b82b56))
* feature flags shows disabled on login until refresh ([14c247b](https://github.com/raghavyuva/nixopus/commit/14c247bb9f3912881a2aa1e1bd44442683d75152))
* fixture loader helps us to create dummy data to the table during development, this sets up the base for creating the development environment with different set of fixtures like testing, development, minimal, complete etc ([11fa6d7](https://github.com/raghavyuva/nixopus/commit/11fa6d7769f72bbe90f055857f8acc5383e7bde5))
* installation message to print out ip:port format ([12f0354](https://github.com/raghavyuva/nixopus/commit/12f0354010e0ef4467e961759b14d9d374afde42))
* make use of users home directory to source the air command after downloading ([67d5644](https://github.com/raghavyuva/nixopus/commit/67d564486932cf992dbd29e54373bbf67e665d8d))
* make view server and api server to run in the background without stopping the program at that point ([7e6a6d4](https://github.com/raghavyuva/nixopus/commit/7e6a6d445f8a5dbf6d406754f67ad51f678a1c9c))
* move to parent directory before starting the view ([ccb3f75](https://github.com/raghavyuva/nixopus/commit/ccb3f75a5419af84ac8f762d6ecc76dae3039110))
* Optional chaining prevents the null pointer error ([f7d9c9b](https://github.com/raghavyuva/nixopus/commit/f7d9c9b3b91d3d7b4f66f85bdc2c52415bbdb5f5))
* permission issues related to air installation, go existance check, and echo statements ([ddd3fdc](https://github.com/raghavyuva/nixopus/commit/ddd3fdce46a07ae815f8fa6a530a0217ed400df7))
* port not displayed after installation with ip based installs ([b88730c](https://github.com/raghavyuva/nixopus/commit/b88730cca224fa7cce7b128c8e5afd8675cfa52d))
* pressing logout from settings page throws null pointer error ([07b68e6](https://github.com/raghavyuva/nixopus/commit/07b68e60c3cc950c34c24a7d232f3f2c9258e2c8))
* prevents non admin users to have the default organization, and only be added to the requested organization through invitation ([0897de5](https://github.com/raghavyuva/nixopus/commit/0897de5518dc1e1be29796a6ab656e4e2fc4ab1c))
* remove asking for confirmation from user when domain validation fails ([0014e84](https://github.com/raghavyuva/nixopus/commit/0014e846972c3a1f9751e91078688c2e55cb11ce))
* remove base from config.mts for documentation site ([33ce717](https://github.com/raghavyuva/nixopus/commit/33ce71709173dd60f8b3773cffa7ec527dedfacc))
* remove checkout to feat/develop branch ([af2eb79](https://github.com/raghavyuva/nixopus/commit/af2eb79f1099d05137fceaf216bf7a1823657f79))
* remove interactive admin credential asking through installation wizard ([cfdb159](https://github.com/raghavyuva/nixopus/commit/cfdb1592bfdce041a64b2706a9c101f3c0885925))
* remove mac-os temporarily ([534a695](https://github.com/raghavyuva/nixopus/commit/534a695078c2d94ae0b3b3b2599382b45aa554e4))
* remove Makefile as it is no longer needed ([8cf8d52](https://github.com/raghavyuva/nixopus/commit/8cf8d526ac8ca07cc266ec10ac494088026fd6f5))
* remove nixopus-staging-redis from the list ([79b4e85](https://github.com/raghavyuva/nixopus/commit/79b4e856b3dac249e3c886fb8f347d37a3a24f9d))
* remove string quotes on parameter passing in qemu steps ([73746af](https://github.com/raghavyuva/nixopus/commit/73746af599526c40b202942e05491c256ccf30f8))
* remove triggering the dev env setup on every pull request and pushes ([c85375d](https://github.com/raghavyuva/nixopus/commit/c85375da27771d9dc8a6294ef5c1896ba9906ccb))
* remove version comparision check ([240716c](https://github.com/raghavyuva/nixopus/commit/240716c4e7dbb95315506c5b1f3b02dbaba97a3d))
* removed docker compose dependency ([d680381](https://github.com/raghavyuva/nixopus/commit/d68038149f73656f4be5fa4a9034e374518c0994))
* removed go installation and auto installation of docker git etc deps as it may cause errors and conflicts ([ed9207f](https://github.com/raghavyuva/nixopus/commit/ed9207fc64088d530cbe756ba7adab7a3180b76b))
* seperate jobs for domain based installation and ip based installation ([b0736ad](https://github.com/raghavyuva/nixopus/commit/b0736ad5fd31c65aeb87d5157a19e282fdcaaeb9))
* unsupported architecture golang install ([e96271c](https://github.com/raghavyuva/nixopus/commit/e96271cfa7bbfd568f508cbae9e75f84493a59e9))
* update complete.yml to use the split imports for different fixtures, and add custom support for importing the fixtures using gopkg/yaml ([01e587f](https://github.com/raghavyuva/nixopus/commit/01e587f8aab667427d298ddec00163f4f455ccd0))
* uses logging module instead of print_debug function for extendability for future changes and to keep consistency ([f709e45](https://github.com/raghavyuva/nixopus/commit/f709e4509befbd71e119a6efd37f339dd927df66))


### Features

* add dev environment setup qemu action ([340e2e3](https://github.com/raghavyuva/nixopus/commit/340e2e3ccd74f66910d6fbe907055874705c40fe))
* Create Issue from dashboard with reporting template and user client infromation in place ([01953f2](https://github.com/raghavyuva/nixopus/commit/01953f2bc47889c3f3cc7f6443a3480752cdd5a8))
* development environment oneclick setup ([81b275a](https://github.com/raghavyuva/nixopus/commit/81b275aeb8fa05f19f61c13b9c9612db44ef29d0))
* include build step for macos ([b15c534](https://github.com/raghavyuva/nixopus/commit/b15c53417b6b35f7d833911c3a2ec215547a2ec1))
* setup script for macos ([b92b04e](https://github.com/raghavyuva/nixopus/commit/b92b04e6b2aaa39d3297d5f3e35bf9e402e012fb))
* Sponsors Showcase on docs ([bb04962](https://github.com/raghavyuva/nixopus/commit/bb04962d9c6f0bba429268a7c6e2d08b08a32aa9))
* Sponsorship Marquee on the Home page ([d7e1211](https://github.com/raghavyuva/nixopus/commit/d7e121198b79648b557c502d7d8800e79a9bdcbb))
* ssh setup logic added for dev setup ([1205995](https://github.com/raghavyuva/nixopus/commit/12059959755cad51d19e8c2675dd096e5fd97526))


### Reverts

* Revert "hot-fix: theming issue due to base path and footer preview in the doc…" (#176) ([39c1aa1](https://github.com/raghavyuva/nixopus/commit/39c1aa107857f92f5917385e9221598be8302436)), closes [#176](https://github.com/raghavyuva/nixopus/issues/176)
* undo changes related to docs ([#195](https://github.com/raghavyuva/nixopus/issues/195)) ([e0b71ec](https://github.com/raghavyuva/nixopus/commit/e0b71ecf75bee61d504523e8c6f6b381dbbeccb3))



# [0.1.0-alpha.8](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.7...v0.1.0-alpha.8) (2025-06-12)


### Bug Fixes

* update version ([dd32047](https://github.com/raghavyuva/nixopus/commit/dd32047d507eae4dca0e1748278e3a71cb8b052e))



# [0.1.0-alpha.7](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.6...v0.1.0-alpha.7) (2025-06-11)


### Bug Fixes

* add title for the introudction blog ([9cb7def](https://github.com/raghavyuva/nixopus/commit/9cb7def97c508504d20c7a20c7a827d034f176ac))
* **container:** listing container fails because of index out of range error (null check issues) ([1fcf064](https://github.com/raghavyuva/nixopus/commit/1fcf064ad760a890c52c35ddf8f247a1e4b4ca7a))
* **docs:** setup node step should find yarn.lock file from docs folder than the root folder ([0b8cad3](https://github.com/raghavyuva/nixopus/commit/0b8cad3e8857d6bdedbdb55c1e7a7ec3c2b1659b))
* does not loop until email and password is provided, rather exits from the installation printing the email is required error ([51d9f4f](https://github.com/raghavyuva/nixopus/commit/51d9f4ffaa4e3ba4f474497b22171a9d42c2da5e))
* fi in the add-sponsors pipeline ([5e19de4](https://github.com/raghavyuva/nixopus/commit/5e19de4f0023893fe6c6ece2801e87a3155a91bd))
* is admin registered data transformation from redux ([da1f66d](https://github.com/raghavyuva/nixopus/commit/da1f66d6c3d606417b27ad341be71396af45152a))
* is_admin controller to return boolean regardless of status ([6c8b5ed](https://github.com/raghavyuva/nixopus/commit/6c8b5edf36bba6e24b2ee6c55c76115d7296aaee))
* **list_containers:** name slicing throws error Index Out of Range ([bb706fa](https://github.com/raghavyuva/nixopus/commit/bb706fa33674fb6f03fc635bb523c6634e62694a))
* localization issues related to registration errors and messages ([1da3bbc](https://github.com/raghavyuva/nixopus/commit/1da3bbc8f72b9426851d5eabd66630dfa5d1f390))
* readme marker for sponsors ([43d5d71](https://github.com/raghavyuva/nixopus/commit/43d5d719a0a331efed4a278b01b9c0e44a79ab52))
* registration requests body to include missing fields ([c57819f](https://github.com/raghavyuva/nixopus/commit/c57819ff3837b90b0cf7f702265aecdf67042275))
* remove broken installation branch from list of triggers in qemu action ([46c72ad](https://github.com/raghavyuva/nixopus/commit/46c72ad06c87866837fa05eeff7dff6bb68f11bc))
* remove custom marker ([e8a930f](https://github.com/raghavyuva/nixopus/commit/e8a930f3bd2a0e40f84aa8b917153aa4b2012a5b))
* replace PAT with GH_TOKEN as secrets in add-sponsors workflow ([2dc4d7a](https://github.com/raghavyuva/nixopus/commit/2dc4d7a8d7a8e924f570b1b19d250eefdf9a6a93))
* service manager and environment.py uses common shared base_config loader ([eac8e26](https://github.com/raghavyuva/nixopus/commit/eac8e268e9ca640074b657094b8b008a33ecfa61))
* **sidebar:** remove container feature from allowed resource in sidebar permission checking ([1cec95d](https://github.com/raghavyuva/nixopus/commit/1cec95d8cf1f5e25e179c1b206e56f648ec02b05))
* **sidebar:** remove container feature from allowed resource in sidebar permission checking ([bf21e58](https://github.com/raghavyuva/nixopus/commit/bf21e586c055d38ce7e52fa5a82592214d756ce7))
* specify docs action to run on every branch pushes, but to be deployed only on master branch ([f121022](https://github.com/raghavyuva/nixopus/commit/f12102288a4f81d8c73ab9563abad40951c154b1))
* sponsor github action ([909d6d3](https://github.com/raghavyuva/nixopus/commit/909d6d3976041eea14b106f5b38dc276f8675ca4))
* standardize password special character validation between generation and validation ([173dca8](https://github.com/raghavyuva/nixopus/commit/173dca8ae280f2dcc39bf66a6c6243b9074790e1))
* syntax issues with docs.yml pushes trigger ([a88ed5f](https://github.com/raghavyuva/nixopus/commit/a88ed5f20f7716e7126c727e9449f978585bce92))
* test input parser uses consistent special chars constant now ([4d0d092](https://github.com/raghavyuva/nixopus/commit/4d0d09207346e6f543aac6f75779d9cb017aa2ae))
* typos in readme.md ([db2c2e4](https://github.com/raghavyuva/nixopus/commit/db2c2e4461e723a3191ad745fa8b990924b7a6f1))
* uses Link tag in loginform for registration navigation ([c7013d3](https://github.com/raghavyuva/nixopus/commit/c7013d3043424e91aca665cade44af3168c9cfe1))
* uses link tag instead of anchor tag, and external links uses security best practices ([5406543](https://github.com/raghavyuva/nixopus/commit/5406543836b67e15783b95d81fa93b9951da68b0))
* uses port decoupled installation script which loads ports and configs from a sepecific config.json file from the helpers/config.json ([f28d520](https://github.com/raghavyuva/nixopus/commit/f28d5200aea74961b85f54aa67062f290c5fec3e))


### Features

* add is-admin-registered api endpoint ([b35722f](https://github.com/raghavyuva/nixopus/commit/b35722ff7f5abac3fa7da13843a52b270dccc4a8))
* add registration page similar to login ui ([47e4d93](https://github.com/raghavyuva/nixopus/commit/47e4d93fb32563fe11dceea15faa7ed464b9c3f1))
* admin credentials are not asked through terminal, rather considers only if provided through arguments ([0ca9f2e](https://github.com/raghavyuva/nixopus/commit/0ca9f2ef12fa1cb39a3a577c3cc9de6f8fb82f8e))
* blogging setup in documentation ([04180f3](https://github.com/raghavyuva/nixopus/commit/04180f3be31f854c909aada35f49e7fa11ae551b))
* checks if admin is registered or not, iff admin not registered then registration screen will be accessible ([37dbb89](https://github.com/raghavyuva/nixopus/commit/37dbb89a07a58dfbb75a4c31eaeed16b31b5c763))
* **coderabbit:** add coderabbit actions and config file ([e584747](https://github.com/raghavyuva/nixopus/commit/e58474744c203f02a8d22b71048c80b7a285cc7d))
* Ip address and port support, no strict domain required ([#131](https://github.com/raghavyuva/nixopus/issues/131)) ([426f06c](https://github.com/raghavyuva/nixopus/commit/426f06c397594a25e730f4d51230834b21a4a82e))
* prevent password exposure as and when user types the password during installation ([dc3f29e](https://github.com/raghavyuva/nixopus/commit/dc3f29eb893cbe300494d6dc9503ac8dea9c8ca6))
* update documentation.md and frontend.md to fix deadlink issues ([87453f8](https://github.com/raghavyuva/nixopus/commit/87453f87ff36ddfa338e65fa84fa8d59c2042815))



# [0.1.0-alpha.5](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.4...v0.1.0-alpha.5) (2025-05-18)


### Bug Fixes

* **caddy:** load caddy file directly instead of traversing and loading the routes ([d60983e](https://github.com/raghavyuva/nixopus/commit/d60983efe021cfee591a1878925c49a6bee74d8b))
* **caddy:** load caddy file directly instead of traversing and loading the routes ([7c900f7](https://github.com/raghavyuva/nixopus/commit/7c900f793e615fe6548d30dc29f46659b07a160b))
* **container:** fails due to missing null pointer checks ([8eb62d2](https://github.com/raghavyuva/nixopus/commit/8eb62d24a12876e825c2787af753617ecfc42397))
* **container:** fails due to missing null pointer checks ([78eb5b2](https://github.com/raghavyuva/nixopus/commit/78eb5b2b7a89d441395461cb8a2ae56745e21b10))
* **decontainer:** uses localWorkspaceFolder instead of /nixopus dir ([82453f0](https://github.com/raghavyuva/nixopus/commit/82453f06bad17f8686ac006416856a2978e9be06))
* **docker_service:** relative path broken finding docker compose file in root dir, now uses absolute path instead ([fb92b6f](https://github.com/raghavyuva/nixopus/commit/fb92b6f4cf2c6dabc581fc8103ea9b7c28c6cbbe))
* **docker-compose:** env path respective to source dir ([0dbd521](https://github.com/raghavyuva/nixopus/commit/0dbd5210e01ae16ca3419776f556fca758109396))
* **docker-compose:** env path respective to source dir ([6b38e8b](https://github.com/raghavyuva/nixopus/commit/6b38e8ba4596b9e78efedd1f07d8697719bb56e1))
* **docker-compose:** env path respective to source dir ([03c93b5](https://github.com/raghavyuva/nixopus/commit/03c93b50a877871c37ef1f276eb943cbc921c7b9))
* **docker-deamon:** overrides default -H fd:// flag from systemd ([c608fe0](https://github.com/raghavyuva/nixopus/commit/c608fe06cbdc4909d6a39d08ffcc12a62a22d4a4))
* **environment-path:** env path according to updated installation script which now has source dir as suffix to nixopus's standard dir ([03b0268](https://github.com/raghavyuva/nixopus/commit/03b02685da0c750eae0db967fe6ba8ed6a4c5e79))
* **go.mod:** update kin-openapi dependency to v0.131.0 ([6e42821](https://github.com/raghavyuva/nixopus/commit/6e42821bde5ec22f1dbe347ae49825416d30d048))
* **installation:** docker tls errors ([1013a97](https://github.com/raghavyuva/nixopus/commit/1013a97a6004f6c3fc728ddb106c90cf45f82305))
* **installer:** docker context creation failure ([5795fe5](https://github.com/raghavyuva/nixopus/commit/5795fe5c64e65def1351d3febdbfe6d2511a3077))
* **installer:** fails to start services docker context inconsistency ([0cacd0c](https://github.com/raghavyuva/nixopus/commit/0cacd0c8b4a53e5135968a00fc24e824eaeedab7))
* **installer:** service manager was using hardcoded 2376 port for connecting to docker ([f280192](https://github.com/raghavyuva/nixopus/commit/f2801922f9cd169de2b80138a057afe58d103317))
* **self-host:** port mapping to match with what caddy listens as a proxy service ([c3a794d](https://github.com/raghavyuva/nixopus/commit/c3a794d4a21057ac9e36fbe31799e07bdd995d58))
* **service_manager:** add debug staatement ([8847fc0](https://github.com/raghavyuva/nixopus/commit/8847fc07bf7aa81f3950ef9a3f3ce41ce823f108))
* **service_manager:** uses etc/nixopus/source instead of /etc/nixopus for source files ([999cde0](https://github.com/raghavyuva/nixopus/commit/999cde03ad2e26dc5876770d707d8002c3214d59))


### Features

* **devcontainer:** restructure Dockerfile and update workspace configuration ([1036ea1](https://github.com/raghavyuva/nixopus/commit/1036ea1dbddc8bf7304328dd8cc5de9c41dd8597))
* **docker:** add installation of air tool in Dockerfile ([52ffa3c](https://github.com/raghavyuva/nixopus/commit/52ffa3ce74d74195ce01ce87c92a6b4b9e8cf856))
* **upload-avatar:** allows users to upload avatar to their account ([eec610b](https://github.com/raghavyuva/nixopus/commit/eec610ba563e6497628b0c71d7a45197c292edac))



# [0.1.0-alpha.4](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.3...v0.1.0-alpha.4) (2025-05-04)


### Bug Fixes

* add common.loading translations to i18n files ([06120ce](https://github.com/raghavyuva/nixopus/commit/06120cee19e4ef67acad520a52660029798f6141))
* add current branch feat/unit_test to test the action ([8f7e831](https://github.com/raghavyuva/nixopus/commit/8f7e831b3182a7be4047c5c752e119c1cd4f8a4d))
* **auto-update:** prevent checking for updates and performing updates in development environment ([916e846](https://github.com/raghavyuva/nixopus/commit/916e846d8d8e579cc0df357b6f89baf34c3e1822))
* **cache:** feature flag middleware throws feature disabled error always ([119919c](https://github.com/raghavyuva/nixopus/commit/119919c32ccc06f66b0807497934c0569938f511))
* **caddy:** proxy caddy json path ([ed66e91](https://github.com/raghavyuva/nixopus/commit/ed66e91689e8b41a288270442c2c4aacd622793d))
* connect to created nixopus user instead of root by default ([8bc60f2](https://github.com/raghavyuva/nixopus/commit/8bc60f2e057f7c97d49c29842c10bdf6bb080891))
* connect to created nixopus user instead of root by default ([f287d77](https://github.com/raghavyuva/nixopus/commit/f287d770a39686486cb75924e76a8a09631ba422))
* **docker-compose-staging:** uses environment as view/.env instead of .env for nixopus-staging-view service ([46cd226](https://github.com/raghavyuva/nixopus/commit/46cd226e3e531a79e1f0f6e668f7ebf9eccb28eb))
* **docker-compose:** env path respective to source dir ([1a40289](https://github.com/raghavyuva/nixopus/commit/1a40289e61c0bc540a2ade808c3cd2c839c4bc81))
* **docker-compose:** env path respective to source dir ([6087943](https://github.com/raghavyuva/nixopus/commit/6087943280a0635023aadbcc90e4a0dcbe78c44b))
* **docker-compose:** env path respective to source dir ([59f63bf](https://github.com/raghavyuva/nixopus/commit/59f63bf9336c8f1c1554c631a88585bff4718a1c))
* **domain:** validation of domain belongs to the server happens only other than development environments ([baa56b7](https://github.com/raghavyuva/nixopus/commit/baa56b78d7cead57e822e939b74670290916237c))
* env field on test action file ([5687388](https://github.com/raghavyuva/nixopus/commit/5687388329775cf384c7c547acf4903a40ad1f35))
* env field on test action file ([99ec34f](https://github.com/raghavyuva/nixopus/commit/99ec34f12239086a9f40d0107b00811ddb1b88de))
* **image-management:** changing filter logic to get the images from docker api ([9ea18b7](https://github.com/raghavyuva/nixopus/commit/9ea18b7e3e4bfe6c9713bc6fdbf26e97e7c2ddc7))
* **installation:** docker tls errors ([f5420b0](https://github.com/raghavyuva/nixopus/commit/f5420b0af200d5a69d34c5fa89782f6fb280b7a0))
* **install:** remove sending output of python script which is a main installer to /dev/null ([58bb6ae](https://github.com/raghavyuva/nixopus/commit/58bb6aec51244b9fc125894126333c75f7f1bf3c))
* **port_confliction:** now randomly assigns a port for the self hosted application, user has to give which port is exposed from the container ([2ebb033](https://github.com/raghavyuva/nixopus/commit/2ebb0338c30af5b5dca8d18036ddb875078525d5))
* **proxy-based-on-environment:** loads caddy config based on environment instead of hardcoding ([83ea802](https://github.com/raghavyuva/nixopus/commit/83ea802d06f0efe45cc6feeedbd09a32a8f22a68))
* **self-host:** port mapping to match with what caddy listens as a proxy service ([69f9d86](https://github.com/raghavyuva/nixopus/commit/69f9d86aa29f59c7862a1076e4dab3ca4e6e43e5))
* **staging-compose:** remove test db service, and change staging docker network to nixopus-staging-network ([7560efb](https://github.com/raghavyuva/nixopus/commit/7560efb2534aa952dd70362a94fae8faa2a289bf))


### Features

* **cache:** adds cache layer for api middleware to cache the context thus by reducing api response time to fewer milliseconds ([fadd646](https://github.com/raghavyuva/nixopus/commit/fadd646514e954705904e130975ccd8f9bc52120))
* **container:** add api endpoints for container management, makes use of existing api/internal/features/deploy/docker/init.go interfaces ([ae73836](https://github.com/raghavyuva/nixopus/commit/ae7383659c14b52222c52668cc3c003cc62bca13))
* **container:** adds image pruning and build cache pruning features through the ui ([3a19009](https://github.com/raghavyuva/nixopus/commit/3a190098c871eb63644e2bf84ec0942ba2b1ce80))
* **containers:** add marketplace ui cards from nixopus's old codebase to container management ([a5872c6](https://github.com/raghavyuva/nixopus/commit/a5872c636bfe34dd031bbdb33616d1993affbf77))
* **container:** wrap the container feature inside feature based access and permission based access logics ([359f55e](https://github.com/raghavyuva/nixopus/commit/359f55e4bd9c00a561d6e25aa4086fcaa386a03c))
* **docker-image-management:** adds endpoint about pruning the docker images, build cache prune, along with list of images retrieval based on filters ([527e64f](https://github.com/raghavyuva/nixopus/commit/527e64f2e2545447eac21e1fc245d47f0de41df2))
* **feature_based_access:** add feature flags components and components to general settings tab ([1a19c0a](https://github.com/raghavyuva/nixopus/commit/1a19c0a7014f76a6f71e6b2f77fcd233e6379607))
* **feature_based_access:** add feature flags database schema and types ([5cc9575](https://github.com/raghavyuva/nixopus/commit/5cc957520b24f8032112a797bdf09e0b7a24b6c5))
* **feature_based_access:** add frontend feature flags redux, setup context provider and types ([e39cba0](https://github.com/raghavyuva/nixopus/commit/e39cba0f50284efaa8018c143d6ba49b9a34f9da))
* **feature_based_access:** implement feature flags controller and core functionality ([1b407f6](https://github.com/raghavyuva/nixopus/commit/1b407f66d1be8589a27a7a51bcbb3c91b720225c))
* **feature_based_access:** integrate feature flags with all the features to restrict in view when disabled ([17d7be5](https://github.com/raghavyuva/nixopus/commit/17d7be53967dc4a8722eeaa280b09d29f3050da8))
* **image-management:** listing of images band integrating with view with styles andcomponents under each container ([278c870](https://github.com/raghavyuva/nixopus/commit/278c870784688774980ca3a1f25f6929478e0e0d))
* **self-host:** allows static file and dockerfile deployment differentiation while showing the form for deployment and configuration ([c513446](https://github.com/raghavyuva/nixopus/commit/c513446ea3fa1da38eb6a6b68a09336f90e63551))


### Performance Improvements

* **cache:** caching for feature flags, so every request will use the cache aside pattern  thus by decreasing the storage lookup time ([9fee21a](https://github.com/raghavyuva/nixopus/commit/9fee21a25b3a464aa51dee6cbd7fbc7bb7e66fd9))



# [0.1.0-alpha.3](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.2...v0.1.0-alpha.3) (2025-04-19)


### Features

* **installation-script:** the bash script that will clone nixopus and runs our installer python package ([e75632b](https://github.com/raghavyuva/nixopus/commit/e75632b5eaefe42fc472c43eafc20559033240bc))
* update installer scripts and validation ([b354087](https://github.com/raghavyuva/nixopus/commit/b354087ea3ea5df9bacd19472aec95d4ae5ce1aa))



# [0.1.0-alpha.2](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.1...v0.1.0-alpha.2) (2025-04-18)


### Bug Fixes

* add required permissions for release-drafter ([0bd91a2](https://github.com/raghavyuva/nixopus/commit/0bd91a2b16759b432f0f8d9afb06a17200b6e427))
* correct sort-by value to merged_at ([db19ec9](https://github.com/raghavyuva/nixopus/commit/db19ec9559e8d9d05547670e428aa3848514976f))
* **openapi-spec:** routes.go to implement consistent grouping strategy for proper openapi spec generation ([804040d](https://github.com/raghavyuva/nixopus/commit/804040d9715fc25820e174ff2b76eb012ca640fd))


### Features

* **update-readme:** add release status badge ([3586c20](https://github.com/raghavyuva/nixopus/commit/3586c201028b2c73df1c8926ea41eeab56ca8c6b))



# [0.1.0-alpha.1](https://github.com/raghavyuva/nixopus/compare/93187f3e6b34a5df7e7d5677ef64dec4608bcd2c...v0.1.0-alpha.1) (2025-04-18)


### Bug Fixes

* **domain-validation:** allow wildcard domain and check only for main domain instead of looking out for *.example.tld in net.LookupIP() ([c12b377](https://github.com/raghavyuva/nixopus/commit/c12b377d277f42d96a3c31dae0cb273b9a443f0a))
* handle missing issue number in release notification ([24a97b3](https://github.com/raghavyuva/nixopus/commit/24a97b389d4043842d3d1ec5b6d5f1ca592ae053))
* **installation-script:** admin registration throws 400 bad request always and not handled properly in our install.py ([d9db6ac](https://github.com/raghavyuva/nixopus/commit/d9db6acb7dfa1b4d960497be56dad9dec006edc6))
* **middleware:** resolve persistent logout issue, add debug logs, update avatar fallback to use username initials ([4a12290](https://github.com/raghavyuva/nixopus/commit/4a12290b9be5d4a75346957d5b4761bfdfe0e4c5))
* **port-issues-view:** keep port next public port when .env copied to view ([bb8570b](https://github.com/raghavyuva/nixopus/commit/bb8570be4b009547aaeeb9134be7d0612e6006aa))
* **pre-commit:** remove pre commit hook ([6d7a779](https://github.com/raghavyuva/nixopus/commit/6d7a7798241ecb21d3ecb312ce522e87c11e974c))
* **README:** Status Badge for the Container Build ([da309f9](https://github.com/raghavyuva/nixopus/commit/da309f96f3255bda8eeaa2dad07129bf9407f1c0))
* **README:** Status Badge for the security scan ([3f54165](https://github.com/raghavyuva/nixopus/commit/3f541651f848300fe56d84a27021a48ce41d2107))
* **rename-action:** renames container ci cd to package manager in build container action workflow ([0b1d189](https://github.com/raghavyuva/nixopus/commit/0b1d189aba3c385dea3ea274244312651550ef53))
* **rename-action:** renames container ci cd to package manager in build container action workflow ([d5f03a6](https://github.com/raghavyuva/nixopus/commit/d5f03a69c3375827a8b7c9de9c6383620a4a7444))
* **update-labeler-action:** labeler action to have contents: write ([a03ba3c](https://github.com/raghavyuva/nixopus/commit/a03ba3cb66d12cec3394a4a1f54a7cee4413ca1c))
* **update-labeler-action:** labeler action to have correct write permission for issues ([222f261](https://github.com/raghavyuva/nixopus/commit/222f26181c066dd2bdc5d4ab4918a13248b4a35e))
* **workflows:** disabled some which are not actually working out ([8c4a6ca](https://github.com/raghavyuva/nixopus/commit/8c4a6cabd14b23863ac91b13a390ab168a009ba8))


### Features

* :sparkles: Rest endpoints for organization roles and permissions for users ([93187f3](https://github.com/raghavyuva/nixopus/commit/93187f3e6b34a5df7e7d5677ef64dec4608bcd2c))
* **docker-image-optimization:** nextjs image size reduction from 2.8gb to 270mb ([b45dd48](https://github.com/raghavyuva/nixopus/commit/b45dd48f4635bd87b7fff14b3aa1da016f92f93e))
* **file-manager:** improve resposiveness of file manager ([4da64a7](https://github.com/raghavyuva/nixopus/commit/4da64a7fc9e0fd6bd177455ba7fe47a404da265d))
* **file-manager:** update with keyboard shortcuts for copy move delete layout change, show hide hidden files creating new files ([ca5aad6](https://github.com/raghavyuva/nixopus/commit/ca5aad65589de3f7b524e1f410d923d180935c10))
* **format-workflow:** pushes as the commit to the same branch ([2c37474](https://github.com/raghavyuva/nixopus/commit/2c374742a1579f751688f9cc9b2afdb1d005dca9))
* **format-workflow:** the format.yaml now formats pull requests and pushes to the branches ([28f540d](https://github.com/raghavyuva/nixopus/commit/28f540d03c71716d93497702cda1cb788c9be19b))
* **labeler:** action that labels our pull requests based on the files changed config specified in labeler.yml ([b8b76a6](https://github.com/raghavyuva/nixopus/commit/b8b76a6f1523f1055897213850b90106f78dabd7))
* **notification:** integrates discord and slack along with email, creates migrations, ui, and controllers and service files to add update delete the webhooks configs ([2bc691e](https://github.com/raghavyuva/nixopus/commit/2bc691eae4051a5805e4868505c2858e3412d656))
* **release-workflow:** debug release workflow ([f3f7a0d](https://github.com/raghavyuva/nixopus/commit/f3f7a0df9065c3ad1c97846f27b61e58fbd3aec9))
* **release-workflow:** debug release workflow ([6551769](https://github.com/raghavyuva/nixopus/commit/655176948365d90e4c4bab4e4243fe0238e9bd59))
* **release-workflow:** debug release workflow ([0c16aff](https://github.com/raghavyuva/nixopus/commit/0c16aff893562a78f6c437aad9a01d6f62f3b782))
* **release-workflow:** debug release workflow ([5006d67](https://github.com/raghavyuva/nixopus/commit/5006d67f9e92cba77f94d5f6097f4a5881e34980))
* **terminal:** fixes issues with terminal writing with spaces, terminal initializing terminal styling issues ([8a67e6b](https://github.com/raghavyuva/nixopus/commit/8a67e6bb9af5fd7bd8b200c2000a392f5378aca7))
* **update-labeler:** labeler uses the PAT instead of access token ([0956d74](https://github.com/raghavyuva/nixopus/commit/0956d74ab9f9c2bba02b2b741332c580e62dc6f7))
* **update-nixopus:** routes for checking for updates / auto updates, and force update of the nixopus app itself todo (implement the service layer for how do we compare the docker image versions and update ([48af332](https://github.com/raghavyuva/nixopus/commit/48af3324bd098992df515430eeb63080c5f2e324))
* **user-settings:** user settings are no more stored in localstorage, it now uses database for patching individual preference like language font etc, user can toggle to choose auto update of nixopus ([98231ad](https://github.com/raghavyuva/nixopus/commit/98231ad1869b801b0047721d5614f288bd6c7112))
* **vulnerability:** fixes CVE-2024-21538 (HIGH) and CVE-2025-30204 (HIGH) ([c25e0c7](https://github.com/raghavyuva/nixopus/commit/c25e0c7fa9ad742a25f95ae7e2a780a881cad573))



