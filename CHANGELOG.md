# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
