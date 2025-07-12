# Audio Note LLM

A desktop application for configuring and processing audio notes using Large Language Models. The application provides a user-friendly interface for selecting audio files, choosing processing actions, and editing prompt templates.

## Overview

Audio Note LLM is a Go-based desktop application built with Fyne that allows users to:
- Select audio files for processing
- Choose from various AI processing actions (blog, paper, requirements, call-to-action)
- Edit and customize prompt templates
- Configure language settings
- Maintain persistent user preferences
- Process audio files with customizable AI prompts

The application features a split-panel interface with configuration controls on the left and a maximized prompt editor on the right, providing an efficient workflow for audio note processing.

# Functions

## Configuration Management

### Dynamic Action Type Loading
- Automatically scans the `config` directory for `prompt-*.txt` files
- Extracts action types from filenames (e.g., `prompt-paper.txt` → `paper`)
- Populates the action type dropdown with discovered options
- Supports extensible prompt templates by simply adding new files

### Persistent Settings
- **Last Action Type**: Remembers and restores the previously selected action
- **Last Language**: Maintains language preference across sessions
- **Last Directory**: Stores the last used directory for file selection
- **Auto-save**: Configuration automatically saves on changes and application exit
- Uses YAML format for human-readable configuration storage

### Smart Defaults
- Defaults to user's Documents directory on first run
- Falls back gracefully when stored directories don't exist
- Automatically selects the last used action type and language
- Provides sensible fallbacks when configuration is missing

## Audio File Management

### File Selection
- File dialog with audio format filtering (MP3 and M4A only)
- Visual feedback showing selected filename
- Directory persistence - dialog opens in last used location
- Hybrid directory location approach for maximum compatibility

### Directory Navigation
- **Smart Starting Location**: File dialog starts in the last used directory
- **Directory Persistence**: Remembers and displays current directory
- **Automatic Updates**: Directory is saved when new files are selected
- **Visual Feedback**: Shows current directory path in the interface

### File Validation
- Validates file selection before processing
- Provides clear error messages for missing files
- Supports common audio formats with focused filtering

## Prompt Editor

### Content Management
- **Dynamic Loading**: Automatically loads prompt content when action type changes
- **Real-time Editing**: Large, maximized text editor for prompt customization
- **Word Wrapping**: Enabled for better readability of long prompts
- **Scrollable Interface**: Handles prompts of any length

### Save Functionality
- **Direct File Writing**: Saves changes directly to prompt files
- **Success Feedback**: Confirmation dialogs for successful saves
- **Error Handling**: Clear error messages for save failures
- **File Preservation**: Maintains file permissions and structure

### Template System
- **Extensible Templates**: Add new action types by creating prompt files
- **Structured Formats**: Each template defines specific output formats
- **Customizable Content**: Full editing capability for all prompt templates

## User Interface

### Split Panel Layout
- **Left Panel**: Configuration controls and file selection
- **Right Panel**: Maximized prompt editor
- **Equal Split**: 50/50 layout for balanced workspace
- **Responsive Design**: Adapts to window resizing

### Visual Design
- **Card Layout**: Organized sections with clear titles and descriptions
- **Bold Labels**: Clear identification of each component
- **Visual Separators**: Clean separation between functional areas
- **Professional Styling**: Consistent theme and spacing

### Progress Tracking
- **Visual Progress Bar**: Shows processing status
- **Button State Management**: Disables controls during processing
- **Status Feedback**: Console output for debugging and monitoring

## Language Support

### Multi-language Configuration
- Support for English (en-US) and German (de-DE)
- Language preference persistence
- Easy extension for additional languages

## Processing Workflow

### Start Process
- **Validation**: Ensures all required selections are made
- **Configuration Save**: Automatically saves current settings
- **Progress Simulation**: Visual feedback during processing
- **Status Updates**: Console logging for process monitoring

### Error Handling
- **Graceful Fallbacks**: Handles missing files and directories
- **User Feedback**: Clear error messages and success confirmations
- **Recovery Options**: Automatic restoration of working directories

## Technical Architecture

### Package Structure
- **Main Application**: GUI logic and user interface
- **Configuration Package**: All config-related functionality separated
- **Clean Architecture**: Separation of concerns for maintainability

### File System Integration
- **Cross-platform Compatibility**: Works on macOS, Windows, and Linux
- **Path Handling**: Proper absolute path resolution
- **Directory Management**: Safe directory changes with restoration

### Data Persistence
- **YAML Configuration**: Human-readable settings storage
- **File-based Templates**: Easy template management and backup
- **Atomic Operations**: Safe file writing with error recovery
