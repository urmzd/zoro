# Changelog

## 0.8.0 (2026-04-05)

### Features

- **chat**: add JSON output flag ([fca5867](https://github.com/urmzd/zoro/commit/fca5867c3ab8d0e66d81a62f89abadb885126fba))
- **cli**: add knowledge graph search and store commands ([a6f5a0a](https://github.com/urmzd/zoro/commit/a6f5a0a1a1d6971f925b66f7fc595816c02fe227))

### Documentation

- expand zoro skill documentation and add preflight script ([0b131ba](https://github.com/urmzd/zoro/commit/0b131ba17e0a318c9f4673a0371dd0033c1f173c))

### Miscellaneous

- add linguist overrides to fix language stats (#6) ([634e6b9](https://github.com/urmzd/zoro/commit/634e6b97d05102230deea9162e6db9732e023a02))

[Full Changelog](https://github.com/urmzd/zoro/compare/v0.7.2...v0.8.0)


## 0.7.2 (2026-04-04)

### Refactoring

- move CLI commands from root into cmd/ package (#5) ([c270bef](https://github.com/urmzd/zoro/commit/c270befb3bcd2eb1424e37aa256a170587edded1))

### Miscellaneous

- update sr action from v2 to v3 ([3451d42](https://github.com/urmzd/zoro/commit/3451d4205b3bf0060133d6b31a43a9ff92232522))

[Full Changelog](https://github.com/urmzd/zoro/compare/v0.7.1...v0.7.2)


## 0.7.1 (2026-03-30)

### Documentation

- add agent skill following agentskills.io spec ([63296a9](https://github.com/urmzd/zoro/commit/63296a9e905c32381709a35e22d036cc708ca99b))

### Refactoring

- replace HTTP/frontend with MCP server, add CLI and knowledge graph (#4) ([8b41394](https://github.com/urmzd/zoro/commit/8b41394ab3d6ffe8a2c9f795a5bfcf8a3c893c94))

### Miscellaneous

- standardize CI/CD — add sr.yaml, CI gate release, fix triggers, justfile recipes ([ecf050e](https://github.com/urmzd/zoro/commit/ecf050eecd5676df23fbf509dcb40caa7b3d5be0))
- use sr-releaser GitHub App for release workflow (#1) ([89a4c26](https://github.com/urmzd/zoro/commit/89a4c264b9eaa7a9e7dcfba94020cbbb64587b46))
- update semantic-release action to sr@v2 ([245af15](https://github.com/urmzd/zoro/commit/245af15a4d621a7413122073bcd64b596be69730))

[Full Changelog](https://github.com/urmzd/zoro/compare/v0.7.0...v0.7.1)


## 0.4.0 (2026-03-07)

### Features

- **graph**: add relationship labels and directional arrows ([75bd31d](https://github.com/urmzd/zoro/commit/75bd31d23ca8aae6868e42893f9a6348d959c033))
- **sidebar**: add sessions sidebar with previous message loading ([0b4bf6c](https://github.com/urmzd/zoro/commit/0b4bf6c35705234283e124c384e75165f0b82aa8))
- **api**: add session listing types and api functions ([8e47063](https://github.com/urmzd/zoro/commit/8e47063e24767b994e02cd7ad9337625c37b0bed))
- **chat**: render messages with streamdown markdown support ([299122d](https://github.com/urmzd/zoro/commit/299122da53aba5b8bad56defae8ff13aa1ca80d2))
- **agent**: add session listing endpoint with improved intent classification ([5e63b91](https://github.com/urmzd/zoro/commit/5e63b91d23f10edbbae969cbe4720716c7b7e2b2))
- **ollama**: add json schema format support to generate requests ([ff23d36](https://github.com/urmzd/zoro/commit/ff23d36fce9a4c164d82780e4d2b183f0bb64c13))

### Miscellaneous

- **deps**: add streamdown dependency for markdown rendering ([0bc3c80](https://github.com/urmzd/zoro/commit/0bc3c80abdf74e4b6cfeb1dbb60e3da322c48854))
- **search**: limit search engines to google, bing, and wikipedia ([ca535bf](https://github.com/urmzd/zoro/commit/ca535bfc14d2717abbadd08a4d940ba0a5a54fd8))
- **docker**: upgrade go runtime to 1.25 and configure fast model ([cb44947](https://github.com/urmzd/zoro/commit/cb449474e7fb48b821e2528d66f48a5bae82db98))


## 0.3.0 (2026-03-04)

### Features

- **frontend-pages**: integrate chat interface and add dynamic chat page ([589d8a1](https://github.com/urmzd/zoro/commit/589d8a1106e1d10e9688ab5b0b973324a6589f1b))
- **frontend-nav**: add header navigation component ([20d3c8a](https://github.com/urmzd/zoro/commit/20d3c8ac558f6eec5601c14cf9233a908e4d27cd))
- **frontend-components**: add knowledge UI components ([b13d1b1](https://github.com/urmzd/zoro/commit/b13d1b1914aefa4b97652e0bdbca030086a01086))
- **frontend-components**: add core chat UI components ([b202d82](https://github.com/urmzd/zoro/commit/b202d82f3b35070d69674f46d66dd1be24ea6b56))
- **frontend-store**: add chat state management ([8edec9e](https://github.com/urmzd/zoro/commit/8edec9ec6ede7c785722fe938d37c284fe773df6))
- **frontend-hooks**: add chat and knowledge chat stream hooks ([50ded17](https://github.com/urmzd/zoro/commit/50ded176cb7770d85071d4f4fb1b5e71008149a7))
- **frontend-api**: add chat and intent API clients ([94e1ea9](https://github.com/urmzd/zoro/commit/94e1ea9315cfc485a35610774ae4014b1706057c))
- **router**: add chat, intent, and autocomplete API routes ([a0bcc6c](https://github.com/urmzd/zoro/commit/a0bcc6c7c6ffa9f3fdbab9e3dbeeddf1981ed256))
- **handler**: add chat, intent, and autocomplete handlers ([ed28030](https://github.com/urmzd/zoro/commit/ed28030297d0880cf8edf774303310f8cd48602d))
- **service**: implement agent and model registry services ([3b95d1f](https://github.com/urmzd/zoro/commit/3b95d1f055993ee21a59f5ed7d19e6d7ba8faba0))
- **model**: add chat models and SSE event types ([1b914b6](https://github.com/urmzd/zoro/commit/1b914b6afc57670052d817980443a8235ffbab51))
- **config**: add SearXNG URL and OllamaFastModel configuration ([c262f39](https://github.com/urmzd/zoro/commit/c262f392b566d3d24729660863e80051c4be1bfc))

### Documentation

- **api**: add generated API documentation ([491447c](https://github.com/urmzd/zoro/commit/491447c102278b0e846b75f824f06fb70d8c4ea1))
- update README and CONTRIBUTING with privacy-first messaging ([6fe5338](https://github.com/urmzd/zoro/commit/6fe5338f12bea69c55e666707e8a770679ac8e3e))

### Refactoring

- **frontend-graph**: improve knowledge graph component ([916a3b3](https://github.com/urmzd/zoro/commit/916a3b33e2f7f2d01cf7064c864b03d67735eed8))

### Miscellaneous

- **infra**: add SearXNG configuration file ([36666d2](https://github.com/urmzd/zoro/commit/36666d22b98782c5b6345a688cd58c131217ea79))
- **deps**: remove chromedp dependency and add SearXNG integration ([b17d182](https://github.com/urmzd/zoro/commit/b17d182c17c3ae6db931b8a3807b47da757e1312))


## 0.2.0 (2026-03-04)

### Features

- **frontend**: add CI, Biome, Turbo, and generated API client ([93d5beb](https://github.com/urmzd/zoro/commit/93d5bebb2c0bb5648663649043433ceba34a414b))


## 0.1.0 (2026-03-04)

### Features

- **frontend**: build research and knowledge graph interface ([69abe77](https://github.com/urmzd/zoro/commit/69abe77b8fa92d0573259342ec1391b43fb086ae))
- **api**: implement knowledge graph backend ([438fdb8](https://github.com/urmzd/zoro/commit/438fdb8b894ced795cdd4f088aa3ebfa02122932))

### Documentation

- add README, CONTRIBUTING guide, and Apache 2.0 license ([49c8c4d](https://github.com/urmzd/zoro/commit/49c8c4db6e0098495fde572ba8a3e502420ed321))
- add RFC for knowledge graph architecture ([ded124e](https://github.com/urmzd/zoro/commit/ded124e9241740f2e56e720c6b72f8818367781a))

### Miscellaneous

- add semantic-release workflow and configuration ([d5cb7a0](https://github.com/urmzd/zoro/commit/d5cb7a04dc87cabcd10c029689e8541b0f51cac1))
- add development and testing scripts ([db41dd0](https://github.com/urmzd/zoro/commit/db41dd0a6da94a575d6ca77fd31002128fe2cce6))
- set up docker infrastructure ([874634d](https://github.com/urmzd/zoro/commit/874634d3d2d6de3a2b5b0fbde706e750d051b49f))
- initialize project configuration ([f1ccb61](https://github.com/urmzd/zoro/commit/f1ccb614b3816f980fffb3ae3f0bfae9e4cfab20))
