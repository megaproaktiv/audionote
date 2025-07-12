## Features implemented:

1. Action Type Select: A dropdown with 3 options - "summary", "call to action", "criticize"
2. Language Select: A dropdown with 2 options - "en-US", "de-DE"
3. Start Button: A button with a play icon that triggers the process
4. Progress Bar: Shows progress when the start button is clicked

## New Features:

1. Audio File Selector: A button labeled "Select Audio File" that opens a file dialog
2. File Filtering: The dialog filters for common audio file formats (.mp3, .wav, .m4a, .aac, .ogg, .flac)
3. Visual Feedback: Once a file is selected, the button text changes to show the selected filename
4. Validation: The start button now checks if a file is selected before proceeding
5. Enhanced Layout: Added the file selector to the configuration card with proper labeling and separators


## Pleasant Layout Features:

• **Card Layout**: The selects are organized in a card with a title and description
• **Bold Labels**: Clear labels for each component
• **Separators**: Visual separation between sections
• **Padding**: Proper spacing around all elements
• **Centered Button**: The start button is centered at the bottom
• **Default Selections**: Both selects have sensible defaults pre-selected
• **Window Sizing**: Set to a comfortable 400x300 pixels

## Functionality:

• When you select different options, they're printed to the console
• Clicking "Start" will:
  • Print the selected action and language
  • Disable the button during processing
  • Animate the progress bar from 0 to 100%
  • Re-enable the button when complete
  • Print "Process completed!" when done



## New Features Implemented:

### 1. Dynamic Prompt File Loading
• Reads all prompt-*.txt files from the config directory at startup
• Extracts action types from filenames (e.g., prompt-paper.txt → paper)
• Populates the action type select with these dynamic options
• Found action types: blog, call-to-action, paper, requirements

### 2. Viper Configuration Management
• Added Viper dependency for YAML configuration management
• Creates config.yaml file in the config directory
• Stores user preferences persistently

### 3. Persistent Settings
• **Last Action Type**: Remembers and restores the last selected action type
• **Last Language**: Remembers and restores the last selected language
• **Auto-save**: Configuration is saved when:
  • User makes selections
  • Start button is clicked
  • Application window is closed

### 4. Enhanced User Experience
• **Smart Defaults**: Uses saved preferences or falls back to first available option
• **Error Handling**: Graceful fallback if prompt files can't be read
• **Console Feedback**: Shows loaded action types and current selections

## Configuration File Structure:
yaml
last_action_type: paper
last_language: de-DE


## How It Works:

1. Startup: Application scans config directory for prompt-*.txt files
2. Population: Action type dropdown is populated with extracted names
3. Loading: Previous selections are loaded from config.yaml
4. Default Selection: Last used action type is automatically selected
5. Saving: Preferences are saved automatically when changed or on exit

## Test Results:
• ✅ Successfully loads action types: [blog call-to-action paper requirements]
• ✅ Creates and reads config.yaml file
• ✅ Remembers last selections (tested with paper and de-DE)
• ✅ Graceful fallback when no config exists
• ✅ Application builds and runs successfully

The application now dynamically adapts to your prompt files and maintains user preferences across sessions, providing
a much more personalized and flexible experience!

### 1. Directory Persistence
• Added LastDirectory field to the configuration structure
• Stores and retrieves the last directory used for file selection
• Configuration is saved to config.yaml and persists across app restarts

### 2. Smart Directory Starting Location
• File dialog now starts in the last used directory
• Uses os.Chdir() approach to change to the stored directory before opening the dialog
• Restores the original working directory after file selection

### 3. Default to Documents Directory
• On first run, defaults to the user's Documents directory (~/Documents)
• Falls back gracefully if Documents directory doesn't exist

### 4. File Filter Restriction
• File dialog now only shows .mp3 and .m4a files
• Removed other audio formats as requested
• Provides cleaner, more focused file selection

### 5. Visual Directory Display
• Shows current directory in an italic label below the file selector
• Updates in real-time when a new directory is selected
• Provides clear feedback about where files will be opened from

## Updated Configuration Structure:
yaml
last_action_type: paper
last_directory: /tmp/test_audio
last_language: de-DE


## How Directory Navigation Works:

1. Startup: Application loads the last used directory from config
2. Display: Shows current directory in the UI
3. File Dialog: When opened, temporarily changes to the stored directory
4. File Selection: User sees files from the last used location first
5. Directory Update: When a file is selected, the new directory is automatically saved
6. Restoration: Original working directory is restored after dialog closes
