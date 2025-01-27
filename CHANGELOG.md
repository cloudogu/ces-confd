# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
- Upgrade go-version to v1.23.5 to fix CVEs in v1.17.13 which no longer receives updates (#36)
  - Replace the `github.com/coreos/etcd/client`-module with `go.etcd.io/etcd/client/v2` to be compatible with new go versions

## [v0.10.0] - 2024-09-18
### Changed
- Relicense to AGPL-3.0-only

## [v0.9.0] - 2024-01-26

### Added
- Added new configuration (`/config/_global/block_warpmenu_support_category`) for completely blocking the support entries in the warp menu (#31)
- Added new configuration (`/config/_global/allowed_warpmenu_support_entries`) for explicitly allowing support entries in the warp menu (#31)

## [v0.8.2] - 2024-01-18
### Fixed
- Fix release pipeline (#28)

## [v0.8.1] - 2024-01-18

### This release is broken and not available due to a broken release pipeline

### Fixed
- Set "Rewrite" to "nil" if the rewrite rule is empty (#26)
  - This prevents creating unused rewrite rules in nginx

## [v0.8.0](https://github.com/cloudogu/ces-confd/releases/tag/v0.8.0) - 2022-12-12
### Added
- Added attribute to enable / disable buffering for specific dogus (#24)

## [v0.7.0](https://github.com/cloudogu/ces-confd/releases/tag/v0.7.0) - 2022-09-20
### Added
- Added service attribute `rewrite` [#22]
- Exported services are now able to define rewrite rules for nginx [#22]

## [v0.6.0](https://github.com/cloudogu/ces-confd/releases/tag/v0.6.0) - 2022-03-22
### Added
- Added support sources which fill the "Support" Category in the warp menu [#20].
- Implemented a filter with the etcd-key "/config/_global/disabled_warpmenu_support_entries" where one can define a list of support entries that should be NOT renderd in warpmenu  [#20].
- Automatic release flow

## [v0.5.1](https://github.com/cloudogu/ces-confd/releases/tag/v0.5.1) - 2021-02-24
### Fixed
- Introduces wait on fail for the several ces-confd watcher to prevent bloated logs [#18]

## [v0.5.0](https://github.com/cloudogu/ces-confd/releases/tag/v0.5.0) - 2021-01-25
### Added
- Implements the reading and setting of the service attribute `location`. [#14]

## [v0.4.0](https://github.com/cloudogu/ces-confd/releases/tag/v0.4.0) - 2021-01-14
### Added
- Implements the service health status to be used within templating [#12]
- Introduces `ignoreHealth`-Property for the service definition in the configuration.yaml to disable health status for debug reasons [#12]
