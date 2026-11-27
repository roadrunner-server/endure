# Changelog

## [2.7.2](https://github.com/roadrunner-server/endure/compare/v2.7.1...v2.7.2) (2026-11-27)


### Bug Fixes

* **stop:** collect root vertices before launching stop goroutines ([41199e7](https://github.com/roadrunner-server/endure/commit/41199e7e4b77e1e20fc94b1fd0a2e697c48eb1f5))

## 2.7.1 (2026-11-27)

### 🎯 Core

- 🐛 **Shutdown order**: Stop plugins in reverse dependency order so providers stay active until dependent plugins finish shutting down, while independent plugins still stop concurrently. See [PR #236](https://github.com/roadrunner-server/endure/pull/236) and [RoadRunner BUG #2471](https://github.com/roadrunner-server/roadrunner/issues/2471).
