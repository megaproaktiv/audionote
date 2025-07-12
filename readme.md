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
