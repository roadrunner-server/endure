CHANGELOG
=========

v2.4.0 (_.08.2021)
-------------------

## 💔 Internal BC:

- 🔨 Pool, worker interfaces: payload now passed and returned by pointer.

## 👀 New:

- ✏️ Long awaited, reworked `Jobs` plugin with pluggable drivers. Now you can allocate/destroy pipelines in the runtime.
  Drivers included in the initial release: `RabbitMQ (0-9-1)`, `SQS v2`, `beanstalk`, `ephemeral`. [PR](https://github.com/spiral/roadrunner/pull/726)
## 🩹 Fixes:

- 🐛 Fix: fixed bug with waiting goroutines on the internal worker's container channel.

## 📈 Summary:

- RR Milestone [2.4.0](https://github.com/spiral/roadrunner/milestone/29)

beta.23 (07.02.2021)
-------------------
- Fix issue when endure doesn't disable vertices which receives disabled interface dependencies.
- CI split into the Linux, Windows, macOS and Linters yml files.
- Update CI badges in the README.md
