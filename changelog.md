# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]


## [v0.2.0]

### Added
- Copy Result Button for resulttext
- when called from CLI, first AWS environment variables are user for credentials
- Bedrock model is configurable from dialog

### Changed
- refactor: panels in own package
- todo SHOULD: Check bucket after saving config
- todo SHOULD: Check authentication after saving config
- todo OPTIONAL: Let path be "same as input path" in config dialog

### Removed

### Fixed

- open load Dialog now uses saved directory

## [v0.1.0]

### Added
- merged "leons ideas"
- read audio files
- transscribe them
- send them to bedrock with template
- Select audio files for processing
- Choose from various AI processing actions (blog, paper, requirements, call-to-action)
- Edit and customize prompt templates
- Configure language settings
- Maintain persistent user preferences
- Process audio files with customizable AI prompts
