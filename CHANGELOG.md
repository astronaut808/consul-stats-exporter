# Changelog

## [2.2.1] 2023-02-25
### Fixed
- Updated dependencies for fix security vulnerability

## [2.2.0] 2020-02-03
### Added
- Add Consul tokenfile support

### Changed
- Update changelog
- Update flags

## [2.1.0] 2020-02-03
### Added
- Add metric `consul_stats_lan_failed_members_count`
- Add metric `consul_stats_lan_left_members_count`
- Add metric `consul_stats_wan_failed_members_count`
- Add metric `consul_stats_wan_left_members_count`

### Changed
- Update changelog
- Update README
- Update comments

## [2.0.0] 2020-02-02
### Added
- Add metric `consul_stats_wan_members_count`
- Add metric `consul_stats_services_count`

### Changed
- Update metric `consul_stats_lan_members_count`
- Update README
- Split `main.go` into multiple files

## [1.2.0] 2020-01-30
### Added
- Add metric `consul_stats_members_count`
- Add metric `consul_stats_bootstrap_expect`
- Add changelog

### Changed
- Update labels for metric `consul_stats_info`
- Update readme

## [1.1.0] 2020-01-29
### Added 
- Add metric `consul_stats_info`

### Changed
- Update readme

## [1.0.0] 2020-01-23
### Added
- Create exporter
- Add readme
