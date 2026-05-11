# [0.1.0-alpha.167](https://github.com/nixopus/nixopus/compare/v0.1.0-alpha.165...v0.1.0-alpha.167) (2026-05-11)


### Bug Fixes

* **csp:** comprehensive CSP and permissions policy fixes ([e7e2678](https://github.com/nixopus/nixopus/commit/e7e2678fb5495884e8f74308070574b44efa801d))
* **installer:** pass ADMIN_PASSWORD to auth container on first boot ([#1342](https://github.com/nixopus/nixopus/issues/1342)) ([5cc6211](https://github.com/nixopus/nixopus/commit/5cc6211d06160190ddf105f2af570f195849e21c))
* **installer:** piped install access URL & provider log ([#1316](https://github.com/nixopus/nixopus/issues/1316)) ([4037588](https://github.com/nixopus/nixopus/commit/4037588a03d884556fa1aa4768c71ab4cb17981f))
* **view:** always verify session on auth init ([#1337](https://github.com/nixopus/nixopus/issues/1337)) ([522b3d1](https://github.com/nixopus/nixopus/commit/522b3d175b77b583e6dbbf4cc4343aaae1c11967)), closes [#1338](https://github.com/nixopus/nixopus/issues/1338)
* **view:** update yarn lock file ([a786883](https://github.com/nixopus/nixopus/commit/a7868838a2204912896bd8eb7bb7487464aacc8c))


### Features

* **agent:** migrate from nixopus/agent to api ([#1336](https://github.com/nixopus/nixopus/issues/1336)) ([b62d098](https://github.com/nixopus/nixopus/commit/b62d098e204b85f23bf6e12405e28c28fca507d0))
* **agent:** schedules tasks ([#1339](https://github.com/nixopus/nixopus/issues/1339)) ([78977cd](https://github.com/nixopus/nixopus/commit/78977cda0ec9b87d5a0b45260826d5ec5414d872))
* **api:** pino-style JSON logs ([#1318](https://github.com/nixopus/nixopus/issues/1318)) ([22edab8](https://github.com/nixopus/nixopus/commit/22edab8065fe09cfc68929644018ce026167eaa3))
* **api:** unify error envelope, add security headers, trace request_id ([#1326](https://github.com/nixopus/nixopus/issues/1326)) ([012ce95](https://github.com/nixopus/nixopus/commit/012ce95627e910d17cdd547141452948a5bef699))
* **billing:** add credits API and billing agent ([#1340](https://github.com/nixopus/nixopus/issues/1340)) ([d9f5011](https://github.com/nixopus/nixopus/commit/d9f5011450b7452636947c3467ea1f0c9d3062bf))
* **installer:** install transcript and nixopus report ([#1317](https://github.com/nixopus/nixopus/issues/1317)) ([f89b804](https://github.com/nixopus/nixopus/commit/f89b80495a7df36f3019c01e0299fe6c293ffa44))
* structured logging across API stack ([#1314](https://github.com/nixopus/nixopus/issues/1314)) ([#1315](https://github.com/nixopus/nixopus/issues/1315)) ([67732b3](https://github.com/nixopus/nixopus/commit/67732b3371c9dc7c475fc7a18171a476beb3b148))
* **view:** add error boundaries ([#1327](https://github.com/nixopus/nixopus/issues/1327)) ([809ad35](https://github.com/nixopus/nixopus/commit/809ad357e97999cb15c5639ef969460f0901ddba))
* **view:** add route-level loading.tsx Suspense fallbacks ([#1328](https://github.com/nixopus/nixopus/issues/1328)) ([5df8a6c](https://github.com/nixopus/nixopus/commit/5df8a6cb242d01b82c22aad523a8bd517fc231e4))
* **view:** add security headers to next.config.ts ([#1329](https://github.com/nixopus/nixopus/issues/1329)) ([c250cc5](https://github.com/nixopus/nixopus/commit/c250cc55a6b356b96e5199597866e040405b87e3))


### Performance Improvements

* **view:** add Million.js compiler ([#1343](https://github.com/nixopus/nixopus/issues/1343)) ([b3a2261](https://github.com/nixopus/nixopus/commit/b3a226148078b9c3a619a65a6a49706379ec6dd8))


### Reverts

* Revert "perf(view): add Million.js compiler (#1343)" (#1344) ([25a4286](https://github.com/nixopus/nixopus/commit/25a4286ac0ee75f25db4049e2faf9b6f1cd2e30c)), closes [#1343](https://github.com/nixopus/nixopus/issues/1343) [#1344](https://github.com/nixopus/nixopus/issues/1344)



