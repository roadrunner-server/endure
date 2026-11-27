# Changelog

## [2.7.2](https://github.com/roadrunner-server/endure/compare/v2.7.1...v2.7.2) (2026-11-27)

### 🐛 Bug Fixes

- **Shutdown root launch race**: Collect all initial vertices with zero active dependents in `Endure.stop()` before starting any stop goroutines. This prevents duplicate stop launches and `sync: negative WaitGroup counter` panics when fast dependents finish before the launch loop reaches their provider. See [PR #238](https://github.com/roadrunner-server/endure/pull/238).

## 2.7.1 (2026-11-27)

### 🎯 Core

- 🐛 **Shutdown order**: Stop plugins in reverse dependency order so providers stay active until dependent plugins finish shutting down, while independent plugins still stop concurrently. See [PR #236](https://github.com/roadrunner-server/endure/pull/236) and [RoadRunner BUG #2471](https://github.com/roadrunner-server/roadrunner/issues/2471).
