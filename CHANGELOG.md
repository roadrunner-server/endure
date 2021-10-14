CHANGELOG
=========

v1.1.0 (14.10.2021)
-------------------
- Fix a bug where disabled vertices don't move the LL pointer forward during the shutdown.
- Prevent from calling Init() method on the already initialized vertices.
- Code cleanup.

## 👀 Packages:

- Go updated to 1.17
- Zap logger updated to 1.19
- Spiral/errors updated to 1.0.12
- Github actions moved to Go 1.17

v1.0.3 (19.08.2021)
-------------------

## 👀 Packages:

- Go updated to 1.17
- Zap logger updated to 1.19
- Spiral/errors updated to 1.0.12
- Github actions moved to Go 1.17

beta.23 (07.02.2021)
-------------------
- Fix issue when endure doesn't disable vertices which receives disabled interface dependencies.
- CI split into the Linux, Windows, macOS and Linters yml files.
- Update CI badges in the README.md
