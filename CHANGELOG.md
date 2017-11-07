# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Documentation

### Changed
- transfer: Fix default configuration error
- agent: Fixed a crash when running plugin
- agent: Fixed a misnamed metric
- agent: Fixed bug which newly deployed agent not honoring authorized_keys
- agent: Move TcpExt collecting to plugin

## [1.0.0] - 2017-08-22
### Added
- agent: Rules repository signing mechanism
- agent: Use cgroups to restrict resource consumption
- agent: Self update mechanism
- agent: Support long running plugins
- rules: New `slot-window` stream for combined condition judging
- alarm: Telegram backend

### Changed
- agent: Configuration format switched from json to yaml, with format tunes
- master: Configuration format switched from json to yaml
- master: Fix a crash bug when access HTTP state
- transfer: Configuration format switched from json to yaml, with format tunes
- transfer: gateway functionality integrated.

## [???]
Lost track...
