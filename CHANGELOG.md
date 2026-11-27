# Changelog

## 2.7.1 (2026-11-27)

### 🎯 Core

- 🐛 **Shutdown order**: Stop plugins in reverse dependency order so providers stay active until dependent plugins finish shutting down, while independent plugins still stop concurrently. See [PR #236](https://github.com/roadrunner-server/endure/pull/236) and [RoadRunner BUG #2471](https://github.com/roadrunner-server/roadrunner/issues/2471).
