# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Todo

2025/08/02 08:29:10 Error starting transcription job: operation error Transcribe: StartTranscriptionJob, https response error StatusCode: 400, RequestID: 0fb0a4f6-5163-4cdb-8c5d-44491af38dbd, BadRequestException: The specified S3 bucket isn't in the same region. Make sure the bucket is in the eu-central-1 region and try your request again.


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
