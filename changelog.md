# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2021-06-25
### Added
- Added errors.List{} type that can hold multiple errors.
- Added errors.Group{} to retrieve an error list from asynchronously executed subtasks.
- Added errors.Has(), errors.HasAnyOf(), errors.Multiple().
- Added changelog.
### Changed
- Error wrappers should work correctly with stdlib's errors.Is() and errors.As().
- errors.E() and similar functions that take interface{} panic more when supplied with invalid arguments.

## [1.1.0] - 2021-06-02
### Added
- Added errors.As() function.
- Added test coverage report.
- Added licence.

## [1.0.0] - 2021-09-19
- Initial release

[Unreleased]: https://github.com/w1ck3dg0ph3r/go-errors/compare/v1.1.0...HEAD
[1.2.0]: https://github.com/w1ck3dg0ph3r/go-errors/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/w1ck3dg0ph3r/go-errors/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/w1ck3dg0ph3r/go-errors/releases/tag/v1.0.0