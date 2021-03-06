# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.5.1](https://github.com/cloudogu/cesapp/releases/tag/v0.5.1) - 2021-02-24
### Fixed
- Introduces wait on fail for the several ces-confd watcher to prevent bloated logs [#18]

## [v0.5.0](https://github.com/cloudogu/cesapp/releases/tag/v0.5.0) - 2021-01-25
### Added
- Implements the reading and setting of the service attribute `location`. [#14]

## [v0.4.0](https://github.com/cloudogu/cesapp/releases/tag/v0.4.0) - 2021-01-14
### Added
- Implements the service health status to be used within templating [#12]
- Introduces `ignoreHealth`-Property for the service definition in the configuration.yaml to disable health status for debug reasons [#12]
