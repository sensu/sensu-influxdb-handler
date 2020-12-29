# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.6.0] - 2020-12-29
### Added
- Added support for legacy functionality via flag `--legacy-format`

### Removed
- Build for darwin/386

## [3.5.0] - 2020-10-20
### Changed
- Updated to latest SDK and Sensu version
- Set password to 'Secret: true' to avoid exposing it in help output

## [3.4.0] - 2020-09-01
### Added
- Added support for an optional boolean flag `--strip-host`.

## [3.3.0] - 2020-06-11
### Changed
- Switched to new community SDK
- Added binary name sensu-influxdb-handler to .gitignore

### Added
- Added config keyspace for annotation support
- Added event annotation for Sensu Enterprise parity

## [3.2.0] - 2020-03-10
### Added
- Added the `--check-status-metric` flag to create metrics from the check status.

### Changed
- Migrated from `dep` to go modules (`go mod`) for managing package dependencies.
- Migrated from TravisCI to GitHub Actions for build, test, and packaging.

## [3.1.2] - 2019-02-21
### Fixed
- Username and password are no longer required to be set, making it possible to
  connect to endpoints that do not have authentication enabled.

### Changed
- Updated travis, goreleaser configurations.
- Updated license.
- If no endpoint addr is provided, the handler will default to http://localhost:8086/

### Removed
- Removed redundant post deploy scripts for travis.

## [3.1.1] - 2019-01-09
### Added
- Adds .bonsai.yml.
- Use of envvar by default for sensitive InfluxDB credentials: addr,username, and password. This prevents leaking of sensitive credential into system process table via command argument. This is a backwards compatible change, commandline arguments can still be used to override envvar values. Here is the envvar to argument mapping:
    - INFLUXDB_ADDR => --addr 
    - INFLUXDB_USER => --username
    - INFLUXDB_PASS => --password

## [3.1.0] - 2018-12-14
### Added
- Adds `--precision` flag (still defaults to 's').
- Adds `--insecure-skip-verify` flag (still defaults to false).

## [3.0.2] - 2018-12-05
### Changed
- Travis post-deploy script generates a sha512 for packages to be sensu asset compatible.

## [3.0.1] - 2018-11-30
### Changed
- Updated the goreleaser file to include env and main in the same
build, hopefully stopping double builds.

## [3.0.0] - 2018-11-30
### Breaking Changes
- Updated sensu-go version to GA RC SHA.
- Updated the goreleaser file so that the handler is packaged as a Sensu
Go compatible asset.

## [v2.0] - 2018-11-21
### Breaking Changes
- Updated sensu-go version to beta-8 and fixed some breaking changes that
were introduced (`Entity.ID` -> `Entity.Name`).
- Changed tag `sensu_entity_id` to `sensu_entity_name` for consistency.

### Removed
- Removed the vendor directory. Dependencies are still managed with Gopkg.toml.

## [v1.8] - 2018-10-23
### Fixed
- Fixed a bug where the handler would only log errors, rather than printing to stderr
and returning a failure exit code.

## [v1.7] - 2018-09-05
### Added
- `Gopkg.lock` and `Gopkg.toml` files

### Changed
- Bumped sensu-go version

## [v1.6] - 2018-08-27
### Added
- Added `grafana-config.json` as a grafana dashboard configuration for the example scripts

### Changed
- Updated help usage for `--addr`

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
