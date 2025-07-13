# Audio Note LLM - Packaging Guide

This document describes how to package and distribute the Audio Note LLM application using the integrated Taskfile commands.

## Prerequisites

- Go 1.19+ installed
- Fyne CLI tool installed (`go install fyne.io/fyne/v2/cmd/fyne@latest`)
- Task runner installed (https://taskfile.dev/)

## Available Commands

### Development Commands

```bash
# Run the application directly
task run

# Development workflow (format, vet, test, run)
task dev

# Format Go code
task fmt

# Run go vet
task vet

# Run tests
task test
```

### Build Commands

```bash
# Build binary only
task build

# Clean build artifacts
task clean

# Show build information
task info
```

### Packaging Commands

```bash
# Package for current platform only (recommended for development)
task package-current

# Package for macOS (creates .app bundle)
task package-darwin

# Package for Windows (creates .exe) - requires Windows or cross-compilation setup
task package-windows

# Package for Linux (creates executable)
task package-linux

# Package for all platforms (may fail on some platforms without proper setup)
task package

# Install macOS app to Applications folder
task install
```

### Release Commands

```bash
# Full release workflow (format, vet, test, package all platforms)
task release
```

## Application Icon

The application uses a custom microphone icon (`icon.png`) that represents the audio processing functionality:

- **Format**: PNG (512x512 pixels)
- **Design**: Material Design-inspired microphone with sound waves
- **Colors**: Blue background with white microphone symbol
- **Usage**: Automatically embedded in packaged applications

## Output Structure

All packaged applications are created in the `dist/` directory:

```
dist/
├── Audio Note LLM.app     # macOS app bundle
├── Audio Note LLM.exe     # Windows executable (if cross-compiled)
└── Audio Note LLM         # Linux executable (if cross-compiled)
```

## Platform-Specific Notes

### macOS
- Creates a proper `.app` bundle with embedded icon
- Can be installed directly to Applications folder using `task install`
- Supports macOS-specific features like app metadata

### Windows
- Creates a `.exe` file with embedded icon
- Requires Windows or proper cross-compilation setup
- May need additional Windows-specific dependencies

### Linux
- Creates a standard executable
- Icon is embedded for desktop environments that support it
- Works on most Linux distributions

## App Metadata

The packaged applications include the following metadata:

- **App Name**: Audio Note LLM
- **App ID**: com.megaproaktiv.audionote-llm
- **Version**: 1.0.0
- **Icon**: Custom microphone icon
- **Description**: Desktop application for audio note processing with LLM

## Troubleshooting

### Common Issues

1. **"fyne command not found"**
   ```bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

2. **Cross-compilation errors**
   - Use `task package-current` for current platform only
   - Set up proper cross-compilation toolchain for other platforms

3. **Icon not found**
   - Ensure `icon.png` exists in the project root
   - Check file permissions

4. **Permission denied on install**
   - Use `sudo task install` if needed
   - Check Applications folder permissions

### Build Information

Use `task info` to check your build environment:

```bash
task info
```

This will show:
- App configuration
- Go version
- Fyne version
- Build paths

## Distribution

The packaged applications in the `dist/` directory are ready for distribution:

- **macOS**: Distribute the `.app` bundle (can be zipped)
- **Windows**: Distribute the `.exe` file
- **Linux**: Distribute the executable (may need to set execute permissions)

For professional distribution, consider code signing and notarization (macOS) or digital signatures (Windows).
