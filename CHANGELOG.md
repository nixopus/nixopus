# [0.1.0-alpha.13](https://github.com/raghavyuva/nixopus/compare/v0.1.0-alpha.12...v0.1.0-alpha.13) (2025-08-09)


### Bug Fixes

* update release cli workflow ([38075dd](https://github.com/raghavyuva/nixopus/commit/38075dd853281eb3954964a6d25f01fd160c7686))



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



