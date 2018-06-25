# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Added `grafana-config.json` as a grafana dashboard configuration for the example scripts

## [v1.5] - 2018-06-22
### Added
- Added `sensu_entity_id` tag

## [v1.4] - 2018-06-21
### Added
- Added `CGO_ENABLED=0` to goreleaser build environment

## [v1.3] - 2018-06-21
### Added
- Added `CGO_ENABLED=0` to travis build environment

## [v1.2] - 2018-05-21
### Added
- `metrics-graphite.sh` to examples
- `metrics-influx.sh` to examples
- `metrics-nagios.sh` to examples
- `metrics-opentsdb.sh` to examples
- `metrics-statsd.sh` to examples

### Fixed
- Fixed errata in `README.md` where example handler name was inconsistent
- Fixed bug for StatsD timestamps

## [v1.1] - 2018-05-16
### Added
- `metrics.sh` script to `examples` directory

### Fixed
- Timestamp translating supports 10 digit int64 timestamps

## [v1.0] - 2018-05-14
### Added
- Initial release
