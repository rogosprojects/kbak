# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.11] - 2025-03-28
### :sparkles: New Features
- [`527ea2e`](https://github.com/rogosprojects/kbak/commit/527ea2e39f560aed98f7ab91fbbcab63aaa0b65f) - extend cleaner utility with support for StatefulSet, DaemonSet, ReplicaSet, Job, CronJob, Ingress, PodDisruptionBudget, Role, ClusterRole, RoleBinding, and ClusterRoleBinding *(commit by [@rogosprojects](https://github.com/rogosprojects))*
- [`ead54c5`](https://github.com/rogosprojects/kbak/commit/ead54c54c3173f439e780871acbeb4c85385447d) - add Windows support to release workflow for cross-platform builds *(commit by [@rogosprojects](https://github.com/rogosprojects))*

### :bug: Bug Fixes
- [`1a22a3c`](https://github.com/rogosprojects/kbak/commit/1a22a3cfae0f407b88089da8dca71dde1fc4a957) - change kubectl dry-run mode from client to server for resource validation *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.10] - 2025-03-09
### :sparkles: New Features
- [`085c7b7`](https://github.com/rogosprojects/kbak/commit/085c7b7917956592fa84f0b06f2c52c44a1d243d) - enhance output messages with colored formatting and emojis for better visibility *(commit by [@rogosprojects](https://github.com/rogosprojects))*
- [`e4417b0`](https://github.com/rogosprojects/kbak/commit/e4417b0715d7d3e8519ed2cea72d0a05c13bf2f7) - add current namespace retrieval and all-namespaces backup functionality *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.9] - 2025-03-06
### :sparkles: New Features
- Add resource type filtering with flags for selectively backing up specific resource types
- Allow users to specify one or more resource types to include in the backup (e.g., --configmap --secret)
- Updated documentation with examples for resource type filtering

## [v0.1.8] - 2025-03-03
### :sparkles: New Features
- [`b5ab86d`](https://github.com/rogosprojects/kbak/commit/b5ab86d5eeb06756ad1135303f36582736ec2959) - add tests for client and backup functionality, improve error handling in main *(commit by [@rogosprojects](https://github.com/rogosprojects))*
- [`0ebb1c4`](https://github.com/rogosprojects/kbak/commit/0ebb1c4c4b1d53e2f5723e4b632252752348e7cf) - add ConfigMap handling *(commit by [@rogosprojects](https://github.com/rogosprojects))*
- [`6b89edf`](https://github.com/rogosprojects/kbak/commit/6b89edf5265a19e9db99d4fd6a0d9649eb266e93) - enhance extraction utility with additional Kubernetes resource support and improve build script *(commit by [@rogosprojects](https://github.com/rogosprojects))*

## [v0.1.7] - 2025-03-03
### :bug: Bug Fixes
- [`7dbab63`](https://github.com/rogosprojects/kbak/commit/7dbab634f202d62375caad2aede91152a5af2d80) - update module paths to use full import paths and improve README instructions *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.6] - 2025-03-03
### :bug: Bug Fixes
- [`8960f4e`](https://github.com/rogosprojects/kbak/commit/8960f4ec93e7d03be046847b18de482d59b68770) - project path in release workflow *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.5] - 2025-03-03
### :sparkles: New Features
- [`6919984`](https://github.com/rogosprojects/kbak/commit/69199846e9452f27693626caecd0687614e78c12) - add kbak backup tool with test script and improve Docker support *(commit by [@rogosprojects](https://github.com/rogosprojects))*

### :recycle: Refactors
- [`e15c154`](https://github.com/rogosprojects/kbak/commit/e15c15458885024f595f42dab1b56b7540a5ebbd) - organize with packages *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.4] - 2025-03-03
### :wrench: Chores
- [`8390c53`](https://github.com/rogosprojects/kbak/commit/8390c5340d80c7f325cc60f9547a1f782e17ce61) - simplify ldflags reference *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.3] - 2025-03-02
### :sparkles: New Features
- [`44c1db1`](https://github.com/rogosprojects/kbak/commit/44c1db15a24cf064e6dd6628b14941e2e25bb4e9) - update README with project logo *(commit by [@rogosprojects](https://github.com/rogosprojects))*

### :bug: Bug Fixes
- [`90c1554`](https://github.com/rogosprojects/kbak/commit/90c15546572821605efcbda1396d99559adf982d) - update ldflags to reference main.Version in release workflow *(commit by [@rogosprojects](https://github.com/rogosprojects))*

### :wrench: Chores
- [`54d649e`](https://github.com/rogosprojects/kbak/commit/54d649eb72785df4db488b2265ab41108d980547) - update README *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [v0.1.2] - 2025-03-02
### :sparkles: New Features
- [`ae7d71e`](https://github.com/rogosprojects/kbak/commit/ae7d71e70b08dd0a505c16e1b00d0be9596de6da) - add versioning support to Docker build and application *(commit by [@rogosprojects](https://github.com/rogosprojects))*

[v0.1.2]: https://github.com/rogosprojects/kbak/compare/v0.1.1...v0.1.2
[v0.1.3]: https://github.com/rogosprojects/kbak/compare/v0.1.2...v0.1.3
[v0.1.4]: https://github.com/rogosprojects/kbak/compare/v0.1.3...v0.1.4
[v0.1.5]: https://github.com/rogosprojects/kbak/compare/v0.1.4...v0.1.5
[v0.1.6]: https://github.com/rogosprojects/kbak/compare/v0.1.5...v0.1.6
[v0.1.7]: https://github.com/rogosprojects/kbak/compare/v0.1.6...v0.1.7
[v0.1.8]: https://github.com/rogosprojects/kbak/compare/v0.1.7...v0.1.8
[v0.1.10]: https://github.com/rogosprojects/kbak/compare/v0.1.9...v0.1.10
[v0.1.11]: https://github.com/rogosprojects/kbak/compare/v0.1.10...v0.1.11
