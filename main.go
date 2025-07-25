package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/megaproaktiv/audionote-config/configuration"
	"github.com/megaproaktiv/audionote-config/llm"
	appPanel "github.com/megaproaktiv/audionote-config/panel"
	"github.com/megaproaktiv/audionote-config/translate"
)

// OutputCapture manages stdout redirection to a text widget
type OutputCapture struct {
	originalStdout *os.File
	pipeReader     *os.File
	pipeWriter     *os.File
	textWidget     *widget.Entry
	mutex          sync.Mutex
}

// NewOutputCapture creates a new output capture instance
func NewOutputCapture(textWidget *widget.Entry) (*OutputCapture, error) {
	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	oc := &OutputCapture{
		originalStdout: os.Stdout,
		pipeReader:     r,
		pipeWriter:     w,
		textWidget:     textWidget,
	}

	// Redirect stdout to our pipe
	os.Stdout = w

	// Start reading from the pipe in a goroutine
	go oc.readOutput()

	return oc, nil
}

// readOutput reads from the pipe and updates the text widget
func (oc *OutputCapture) readOutput() {
	buffer := make([]byte, 1024)
	for {
		n, err := oc.pipeReader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from pipe: %v", err)
			}
			break
		}

		if n > 0 {
			// Write to original stdout
			oc.originalStdout.Write(buffer[:n])

			// Update text widget
			text := string(buffer[:n])
			text = strings.TrimRight(text, "\n\r")
			if text != "" {
				// Use fyne.Do to ensure UI updates happen on main thread
				fyne.Do(func() {
					oc.mutex.Lock()
					currentText := oc.textWidget.Text
					if currentText != "" {
						newText := currentText + "\n" + text
						oc.textWidget.SetText(newText)
					} else {
						oc.textWidget.SetText(text)
					}
					// Move cursor to end for auto-scroll effect
					oc.textWidget.CursorRow = len(strings.Split(oc.textWidget.Text, "\n")) - 1
					oc.textWidget.CursorColumn = len(strings.Split(oc.textWidget.Text, "\n")[oc.textWidget.CursorRow])
					oc.mutex.Unlock()
				})
			}
		}
	}
}

// Close restores original stdout and closes the pipe
func (oc *OutputCapture) Close() {
	os.Stdout = oc.originalStdout
	oc.pipeWriter.Close()
	oc.pipeReader.Close()
}

// updateOutputFieldSize updates the output field size based on configuration
func updateOutputFieldSize(outputField *widget.Entry, config *configuration.Config) {
	// Calculate height based on configured output lines (increased per-line height and padding)
	outputHeight := float32(config.OutputLines*22 + 50)
	if outputHeight < 220 { // Minimum height
		outputHeight = 220
	}
	if outputHeight > 400 { // Maximum height for reasonable display
		outputHeight = 400
	}
	outputField.Resize(fyne.NewSize(450, outputHeight))
	fmt.Printf("Output field resized for %d lines (height: %.0f)\n", config.OutputLines, outputHeight)
}

// checkForExistingTranscript checks if a transcript already exists for the given audio file
func checkForExistingTranscript(audioFilePath, bucket, language string) string {
	// Generate the expected job name based on the audio file
	fileName := filepath.Base(audioFilePath)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Check for existing transcript files in the local directory
	// Look for files matching multiple patterns to catch all variations:
	// 1. summary/output/{baseName}-DMIN-*.json (original filename)
	// 2. summary/output/{baseName}_Copy-DMIN-*.json (filename with _Copy suffix)
	// 3. summary/output/{baseName}_copy.{ext}-DMIN-*.json (filename with _copy suffix and extension)
	outputDir := "summary/output"
	pattern1 := filepath.Join(outputDir, baseName+"-DMIN-*.json")
	pattern2 := filepath.Join(outputDir, baseName+"_Copy-DMIN-*.json")
	pattern3 := filepath.Join(outputDir, baseName+"_copy.*-DMIN-*.json")

	fmt.Printf("Searching for existing transcripts with patterns:\n")
	fmt.Printf("  Pattern 1: %s\n", pattern1)
	fmt.Printf("  Pattern 2: %s\n", pattern2)
	fmt.Printf("  Pattern 3: %s\n", pattern3)

	// Search for all patterns
	matches1, err := filepath.Glob(pattern1)
	if err != nil {
		fmt.Printf("Error searching for existing transcripts (pattern 1): %v\n", err)
	}

	matches2, err := filepath.Glob(pattern2)
	if err != nil {
		fmt.Printf("Error searching for existing transcripts (pattern 2): %v\n", err)
	}

	matches3, err := filepath.Glob(pattern3)
	if err != nil {
		fmt.Printf("Error searching for existing transcripts (pattern 3): %v\n", err)
	}

	// Combine all matches
	allMatches := append(matches1, matches2...)
	allMatches = append(allMatches, matches3...)

	// If we find matching files, try to read the most recent one
	if len(allMatches) > 0 {
		fmt.Printf("Found %d matching transcript files\n", len(allMatches))
		// Sort by modification time to get the most recent
		var latestFile string
		var latestTime int64

		for _, match := range allMatches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.ModTime().Unix() > latestTime {
				latestTime = info.ModTime().Unix()
				latestFile = match
			}
		}

		if latestFile != "" {
			fmt.Printf("Found existing transcript file: %s\n", latestFile)
			// Try to read and parse the transcript
			data, err := os.ReadFile(latestFile)
			if err != nil {
				fmt.Printf("Error reading existing transcript: %v\n", err)
				return ""
			}

			// Parse the JSON to extract the transcript text
			var transcriptResp struct {
				Results struct {
					Transcripts []struct {
						Transcript string `json:"transcript"`
					} `json:"transcripts"`
				} `json:"results"`
			}

			if err := json.Unmarshal(data, &transcriptResp); err != nil {
				fmt.Printf("Error parsing existing transcript JSON: %v\n", err)
				return ""
			}

			if len(transcriptResp.Results.Transcripts) > 0 {
				fmt.Printf("Successfully loaded existing transcript (%d characters)\n", len(transcriptResp.Results.Transcripts[0].Transcript))
				return transcriptResp.Results.Transcripts[0].Transcript
			}
		}
	}

	fmt.Printf("No existing transcript found for %s\n", audioFilePath)
	return ""
}

// showAboutDialog displays the About dialog
func showAboutDialog(w fyne.Window) {
	aboutContent := widget.NewRichTextFromMarkdown(`# Audio Note LLM

A desktop application for configuring and processing audio notes using Large Language Models.

## Features
• **Audio File Support**: Process MP3 and M4A files
• **AI Processing Actions**: Choose from various processing templates
• **Prompt Editor**: Edit and customize AI prompt templates
• **Language Support**: Multiple language configurations
• **AWS Integration**: S3 bucket and profile configuration
• **Persistent Settings**: Automatic saving of user preferences

## Technical Details
• Built with **Go** programming language
• UI framework: **Fyne v2**
• Configuration: **Viper** with YAML storage
• Cross-platform compatibility

## Version
**1.0.0** - Initial Release

---
*Built with ❤️ for efficient audio note processing*`)

	aboutDialog := dialog.NewCustom("About Audio Note LLM", "Close", aboutContent, w)
	aboutDialog.Resize(fyne.NewSize(500, 400))
	aboutDialog.Show()
}

// showConfigDialog displays the configuration dialog
func showConfigDialog(w fyne.Window, config *configuration.Config, outputField *widget.Entry) {
	// Create entry widgets for configuration
	s3BucketEntry := widget.NewEntry()
	s3BucketEntry.SetText(config.S3Bucket)
	s3BucketEntry.SetPlaceHolder("Enter S3 bucket name (e.g., my-audio-bucket)")

	awsProfileEntry := widget.NewEntry()
	awsProfileEntry.SetText(config.AWSProfile)
	awsProfileEntry.SetPlaceHolder("Enter AWS profile name (e.g., default)")

	// Create output path entry
	outputPathEntry := widget.NewEntry()
	outputPathEntry.SetText(config.OutputPath)
	outputPathEntry.SetPlaceHolder("Enter output file path (e.g., /path/to/result.txt)")

	// Create output lines slider
	outputLinesSlider := widget.NewSlider(5, 50)
	outputLinesSlider.SetValue(float64(config.OutputLines))
	outputLinesSlider.Step = 1

	outputLinesLabel := widget.NewLabel(fmt.Sprintf("Output Lines: %d", config.OutputLines))
	outputLinesSlider.OnChanged = func(value float64) {
		outputLinesLabel.SetText(fmt.Sprintf("Output Lines: %d", int(value)))
	}

	// Create labels with descriptions
	s3Label := widget.NewRichTextFromMarkdown("**S3 Bucket:**\nThe AWS S3 bucket where audio files will be stored or retrieved.")
	awsLabel := widget.NewRichTextFromMarkdown("**AWS Profile:**\nThe AWS CLI profile to use for authentication.")
	outputPathLabel := widget.NewRichTextFromMarkdown("**Output File Path:**\nThe path where the processing result will be saved.")
	outputLabel := widget.NewRichTextFromMarkdown("**Output Display Lines:**\nMinimum number of lines to display in the output area (5-50).")

	// Create form content
	formContent := container.NewVBox(
		s3Label,
		s3BucketEntry,
		widget.NewSeparator(),
		awsLabel,
		awsProfileEntry,
		widget.NewSeparator(),
		outputPathLabel,
		outputPathEntry,
		widget.NewSeparator(),
		outputLabel,
		outputLinesLabel,
		outputLinesSlider,
		widget.NewSeparator(),
		widget.NewLabel("Note: Make sure your AWS credentials are properly configured."),
	)

	// Create dialog
	configDialog := dialog.NewCustomConfirm(
		"Configuration Settings",
		"Save",
		"Cancel",
		formContent,
		func(confirmed bool) {
			if confirmed {
				// Basic validation
				s3Bucket := s3BucketEntry.Text
				awsProfile := awsProfileEntry.Text
				outputPath := outputPathEntry.Text
				outputLines := int(outputLinesSlider.Value)

				if awsProfile == "" {
					awsProfile = "default"
				}

				// Validate output path
				if outputPath == "" {
					outputPath = filepath.Join(config.LastDirectory, "result.txt")
				}

				// Validate output lines
				if outputLines < 5 {
					outputLines = 5
				} else if outputLines > 50 {
					outputLines = 50
				}

				// Update configuration
				config.S3Bucket = s3Bucket
				config.AWSProfile = awsProfile
				config.OutputPath = outputPath
				config.OutputLines = outputLines

				// Save configuration
				config.Save()

				// Update output field size if it changed
				if outputField != nil {
					updateOutputFieldSize(outputField, config)
				}

				// Show success message
				successMsg := fmt.Sprintf("Configuration saved successfully!\n\nS3 Bucket: %s\nAWS Profile: %s\nOutput Path: %s\nOutput Lines: %d",
					s3Bucket, awsProfile, outputPath, outputLines)
				dialog.ShowInformation("Configuration Saved", successMsg, w)

				fmt.Printf("Configuration updated - S3 Bucket: %s, AWS Profile: %s, Output Path: %s, Output Lines: %d\n",
					config.S3Bucket, config.AWSProfile, config.OutputPath, config.OutputLines)
			}
		},
		w,
	)

	configDialog.Resize(fyne.NewSize(500, 450))
	configDialog.Show()
}

func main() {
	//--------------------------------------------------------------
	// Initialize application and window
	//--------------------------------------------------------------
	a := app.New()
	w := a.NewWindow("Audio Note LLM")
	w.Resize(fyne.NewSize(1400, 900)) // Increased size to fully show output and new tab layout

	//--------------------------------------------------------------
	// Initialize configuration and context
	//--------------------------------------------------------------
	config := configuration.InitConfig()

	//--------------------------------------------------------------
	// Create output field for stdout capture
	//--------------------------------------------------------------
	outputField := widget.NewMultiLineEntry()
	// to not explode screen
	maxlines := min(config.OutputLines, 20)
	maxlines = max(maxlines, 10)
	outputField.SetMinRowsVisible(maxlines)
	// outputField.Disable() // Make it read-only
	outputField.Wrapping = fyne.TextWrapWord
	outputField.SetPlaceHolder("Application output will appear here...")

	// Style the output field with smaller font
	outputField.TextStyle = fyne.TextStyle{
		Monospace: true, // Use monospace font for better readability
	}

	// Note: Size will be set when creating the scroll container

	//--------------------------------------------------------------
	// Create application menu
	//--------------------------------------------------------------
	aboutMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About Audio Note LLM", func() {
			showAboutDialog(w)
		}),
	)

	configMenu := fyne.NewMenu("Settings",
		fyne.NewMenuItem("Configuration...", func() {
			showConfigDialog(w, config, outputField)
		}),
	)

	mainMenu := fyne.NewMainMenu(configMenu, aboutMenu)
	w.SetMainMenu(mainMenu)

	//--------------------------------------------------------------
	// Load action types from prompt files
	//--------------------------------------------------------------
	actionTypes, err := configuration.LoadPromptFiles()
	if err != nil {
		fmt.Printf("Error loading prompt files: %v\n", err)
		actionTypes = []string{"blog", "paper", "requirements", "call-to-action"} // fallback
	}

	if len(actionTypes) == 0 {
		actionTypes = []string{"blog", "paper", "requirements", "call-to-action"} // fallback
	}

	fmt.Printf("Loaded action types: %v\n", actionTypes)

	//--------------------------------------------------------------
	// Create prompt editor and content loading function
	//--------------------------------------------------------------
	promptEditor := widget.NewMultiLineEntry()
	promptEditor.Wrapping = fyne.TextWrapWord
	promptEditor.SetPlaceHolder("Select an action type to load its prompt content...")

	// Function to load prompt content
	loadPromptContent := func(actionType string) {
		content, err := configuration.LoadPromptContent(actionType)
		if err != nil {
			promptEditor.SetText(fmt.Sprintf("Error loading prompt for '%s': %v", actionType, err))
			fmt.Printf("Error loading prompt content for %s: %v\n", actionType, err)
		} else {
			promptEditor.SetText(content)
			fmt.Printf("Loaded prompt content for action type: %s\n", actionType)
		}
	}

	//--------------------------------------------------------------
	// Create action type selector
	//--------------------------------------------------------------
	actionSelect := widget.NewSelect(
		actionTypes,
		func(value string) {
			fmt.Printf("Action selected: %s\n", value)
			config.LastActionType = value
			// Load the corresponding prompt content
			loadPromptContent(value)
		},
	)

	// Set default selection from config or first available option
	defaultAction := config.LastActionType
	if defaultAction == "" || !configuration.Contains(actionTypes, defaultAction) {
		defaultAction = actionTypes[0]
	}
	actionSelect.SetSelected(defaultAction)
	config.LastActionType = defaultAction

	//--------------------------------------------------------------
	// Create language selector
	//--------------------------------------------------------------
	languageSelect := widget.NewSelect(
		[]string{"en-US", "de-DE"},
		func(value string) {
			fmt.Printf("Language selected: %s\n", value)
			config.LastLanguage = value
		},
	)

	// Set default language from config
	if config.LastLanguage != "" {
		languageSelect.SetSelected(config.LastLanguage)
	} else {
		languageSelect.SetSelected("en-US")
		config.LastLanguage = "en-US"
	}

	//--------------------------------------------------------------
	// Create file selector for audio files
	//--------------------------------------------------------------
	var selectedFilePath string
	var fileSelector *widget.Button
	var directoryLabel *widget.Label

	fileSelector = widget.NewButton("Select Audio File", func() {
		// Store current directory to restore later
		currentDir, _ := os.Getwd()
		fmt.Printf("Current working directory: %s\n", currentDir)

		// Set the directory for the dialog
		if err := config.SetDirectoryForDialog(); err != nil {
			fmt.Printf("Could not set directory: %v\n", err)
		}

		dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Always restore original directory first
			configuration.RestoreDirectory(currentDir)

			if err != nil {
				fmt.Printf("Error selecting file: %v\n", err)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			selectedFilePath = reader.URI().Path()
			fileSelector.SetText(fmt.Sprintf("Selected: %s", filepath.Base(selectedFilePath)))
			fmt.Printf("Audio file selected: %s\n", selectedFilePath)

			// Update last used directory
			config.LastDirectory = filepath.Dir(selectedFilePath)
			directoryLabel.SetText(fmt.Sprintf("Directory: %s", config.LastDirectory))
			fmt.Printf("Updated last directory to: %s\n", config.LastDirectory)
		}, w)

		// Set file filter for only m4a and mp3 files
		dialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3", ".m4a"}))

		// Also try to set location via URI (additional method)
		if dirURI := config.GetDirectoryURI(); dirURI != nil {
			// Try to cast to ListableURI for SetLocation
			if listableURI, ok := dirURI.(fyne.ListableURI); ok {
				dialog.SetLocation(listableURI)
				fmt.Printf("Also set dialog URI location to: %s\n", config.LastDirectory)
			} else {
				fmt.Printf("URI is not listable, relying on directory change method\n")
			}
		}

		fmt.Printf("Opening file dialog (should start in: %s)\n", config.LastDirectory)
		dialog.Show()
	})

	//--------------------------------------------------------------
	// Create output path selector
	//--------------------------------------------------------------
	var outputPathSelector *widget.Button
	var outputDirectoryLabel *widget.Label
	outputPathSelector = widget.NewButton("Select Output Path", func() {
		// Store current directory to restore later
		currentDir, _ := os.Getwd()
		fmt.Printf("Current working directory: %s\n", currentDir)

		// Set the directory for the dialog (use the directory of current output path if it exists)
		outputDir := filepath.Dir(config.OutputPath)
		if outputDir != "" && configuration.DirExists(outputDir) {
			if err := os.Chdir(outputDir); err != nil {
				fmt.Printf("Could not set output directory: %v\n", err)
			}
		}
		panel := appPanel.Panel{
			CurrentDir:           currentDir,
			OutputPathSelector:   outputPathSelector,
			OutputDirectoryLabel: outputDirectoryLabel,
		}
		_, err := appPanel.SelectButton(
			&panel, config)

		if err != nil {
			fmt.Printf("Error opening file dialog: %v\n", err)
			return
		}

	})

	//--------------------------------------------------------------
	// Create progress bar
	//--------------------------------------------------------------
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0.0)

	//--------------------------------------------------------------
	// Set up stdout capture
	//--------------------------------------------------------------
	outputCapture, err := NewOutputCapture(outputField)
	if err != nil {
		fmt.Printf("Error setting up output capture: %v\n", err)
	}

	// Add initial message to output
	fmt.Println("Audio Note LLM started - output will be displayed here")

	//--------------------------------------------------------------
	// Create action management functions
	//--------------------------------------------------------------
	// Function to refresh action types from directory
	refreshActionTypes := func() {
		newActionTypes, err := configuration.LoadPromptFiles()
		if err != nil {
			fmt.Printf("Error loading prompt files: %v\n", err)
			return
		}

		if len(newActionTypes) == 0 {
			newActionTypes = []string{"blog", "paper", "requirements", "call-to-action"} // fallback
		}

		// Update the select widget options
		actionSelect.Options = newActionTypes

		// If current selection is no longer valid, select the first option
		if !configuration.Contains(newActionTypes, actionSelect.Selected) {
			if len(newActionTypes) > 0 {
				actionSelect.SetSelected(newActionTypes[0])
				config.LastActionType = newActionTypes[0]
				loadPromptContent(newActionTypes[0])
			}
		}

		fmt.Printf("Refreshed action types: %v\n", newActionTypes)
	}

	//--------------------------------------------------------------
	// Create prompt management buttons
	//--------------------------------------------------------------
	// Create save button for prompt editor
	savePromptButton := widget.NewButtonWithIcon("Save Prompt", theme.DocumentSaveIcon(), func() {
		currentAction := actionSelect.Selected
		if currentAction == "" {
			dialog.ShowError(fmt.Errorf("no action type selected"), w)
			return
		}

		content := promptEditor.Text
		err := configuration.SavePromptContent(currentAction, content)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save prompt: %v", err), w)
			fmt.Printf("Error saving prompt for %s: %v\n", currentAction, err)
		} else {
			dialog.ShowInformation("Success", fmt.Sprintf("Prompt for '%s' saved successfully!", currentAction), w)
			fmt.Printf("Successfully saved prompt for action type: %s\n", currentAction)
		}
	})

	// Create refresh button for action types
	refreshActionButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		refreshActionTypes()
		dialog.ShowInformation("Refreshed", "Action types refreshed from directory", w)
	})

	// Create new action button
	newActionButton := widget.NewButtonWithIcon("New Action", theme.ContentAddIcon(), func() {
		// Create a simple dialog to get the new action name
		actionNameEntry := widget.NewEntry()
		actionNameEntry.SetPlaceHolder("Enter action name (e.g., summary, analysis, notes)")

		// Validate action name (no spaces, only alphanumeric and hyphens)
		actionNameEntry.OnChanged = func(text string) {
			// Remove spaces and special characters
			cleanText := strings.ReplaceAll(text, " ", "-")
			cleanText = strings.ToLower(cleanText)
			// Keep only alphanumeric and hyphens
			var result strings.Builder
			for _, char := range cleanText {
				if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
					result.WriteRune(char)
				}
			}
			if result.String() != text {
				actionNameEntry.SetText(result.String())
			}
		}

		content := container.NewVBox(
			widget.NewLabel("Create a new action type:"),
			actionNameEntry,
			widget.NewLabel("This will create a new prompt template file."),
		)

		confirmDialog := dialog.NewCustomConfirm(
			"New Action",
			"Create",
			"Cancel",
			content,
			func(confirmed bool) {
				if confirmed {
					actionName := actionNameEntry.Text
					if actionName == "" {
						dialog.ShowError(fmt.Errorf("action name cannot be empty"), w)
						return
					}

					// Create a default prompt template
					defaultPrompt := fmt.Sprintf(`# %s Template

Write your own prompt.
The save the prompt before you want to use it.

Please process the following audio transcript and create a %s:

## Instructions:
- Analyze the content thoroughly
- Structure the output clearly
- Provide actionable insights


## %s:`, strings.Title(actionName), actionName, strings.Title(actionName))

					// Save the new prompt
					err := configuration.SavePromptContent(actionName, defaultPrompt)
					if err != nil {
						dialog.ShowError(fmt.Errorf("failed to create new action: %v", err), w)
						fmt.Printf("Error creating new action %s: %v\n", actionName, err)
					} else {
						// Refresh the action types and select the new one
						refreshActionTypes()
						actionSelect.SetSelected(actionName)
						config.LastActionType = actionName
						loadPromptContent(actionName)

						dialog.ShowInformation("Success", fmt.Sprintf("New action '%s' created successfully!", actionName), w)
						fmt.Printf("Successfully created new action: %s\n", actionName)
					}
				}
			},
			w,
		)

		confirmDialog.Resize(fyne.NewSize(400, 200))
		confirmDialog.Show()
	})

	// Prompt editor label
	promptLabel := widget.NewLabel("Prompt Editor:")
	promptLabel.TextStyle.Bold = true

	//--------------------------------------------------------------
	// Create result display field
	//--------------------------------------------------------------
	resultField := widget.NewMultiLineEntry()
	//resultField.Disable() // Make it read-only
	resultField.Wrapping = fyne.TextWrapWord
	resultField.SetPlaceHolder("Processing results will appear here...")
	resultField.TextStyle = fyne.TextStyle{
		Monospace: false, // Use regular font for results
	}

	// Load existing result file if it exists
	if resultContent, err := os.ReadFile(config.OutputPath); err == nil {
		resultField.SetText(string(resultContent))
		fmt.Printf("Loaded existing result from %s\n", config.OutputPath)
	}

	//--------------------------------------------------------------
	// Create right panel with tabs
	//--------------------------------------------------------------
	rightPanel := appPanel.RightPanel(
		appPanel.Panel{
			PromptLabel:      promptLabel,
			PromptEditor:     promptEditor,
			ResultField:      resultField,
			SavePromptButton: savePromptButton,
		},
	)

	//--------------------------------------------------------------
	// Create start button and processing logic
	//--------------------------------------------------------------
	// Create the start button with Material Design microphone icon
	// Using emoji + built-in icon for better compatibility
	var startButton *widget.Button
	startButton = widget.NewButtonWithIcon("🎤 Start", theme.VolumeUpIcon(), func() {
		action := actionSelect.Selected
		language := languageSelect.Selected

		if selectedFilePath == "" {
			dialog.ShowInformation("No File Selected", "Please select an audio file first.", w)
			return
		}

		// Save current configuration
		config.Save()

		fmt.Printf("Starting process with Action: %s, Language: %s, File: %s\n", action, language, selectedFilePath)

		//--------------------------------------------------------------
		// Start processing
		//--------------------------------------------------------------
		go func() {
			fyne.Do(func() {
				startButton.Disable()
				progressBar.SetValue(float64(10) / 100.0)
			})
			ctx := context.Background()
			// Check if transcript already exists
			fmt.Printf("Checking for existing transcript...\n")
			existingTranscript := checkForExistingTranscript(selectedFilePath, config.S3Bucket, language)

			var transcript string
			if existingTranscript != "" {
				fmt.Printf("Found existing transcript, skipping transcription process\n")
				transcript = existingTranscript
				fyne.Do(func() {
					progressBar.SetValue(float64(50) / 100.0)
				})
			} else {
				fmt.Printf("No existing transcript found, starting transcription with language: %s\n", language)
				fyne.Do(func() {
					progressBar.SetValue(float64(20) / 100.0)
				})
				awsProfile := config.AWSProfile
				err = translate.InitClient(ctx, awsProfile)
				if err != nil {
					fyne.Do(func() {
						progressBar.SetValue(float64(0.0))
						fmt.Println("Could not load AWS profile: ", awsProfile)
						startButton.Enable()
					})
					return
				}
				fyne.Do(func() {
					progressBar.SetValue(float64(30) / 100.0)
				})
				transcript = translate.Translate(ctx, translate.Client, selectedFilePath, config.S3Bucket, language)
				fyne.Do(func() {
					progressBar.SetValue(float64(50) / 100.0)
				})

			}
			promptData, err := configuration.LoadPromptContent(action)
			if err != nil {
				log.Fatalf("Error loading prompt: %v", err)
			}
			fullPrompt := string(promptData) + "\n" + transcript

			bedrockResult, err := llm.CallBedrock(fullPrompt)
			if err != nil {
				log.Fatalf("Error calling Bedrock: %v", err)
			}

			err = os.WriteFile(config.OutputPath, []byte(bedrockResult), 0644)
			if err != nil {
				log.Fatalf("Error writing result to %s: %v", config.OutputPath, err)
			}
			fmt.Printf("Done. Result written to %s\n", config.OutputPath)

			// Load result into the result tab
			resultContent, err := os.ReadFile(config.OutputPath)
			if err != nil {
				fmt.Printf("Error reading result from %s: %v\n", config.OutputPath, err)
				fyne.Do(func() {
					resultField.SetText("Error loading result file")
				})
			} else {
				fyne.Do(func() {
					resultField.SetText(string(resultContent))
					fmt.Println("Result loaded into Result tab")
				})
			}

			// Switch to the Result tab to show the result
			fyne.Do(func() {
				rightPanel.SelectTab(rightPanel.Items[1]) // Switch to second tab (Result)
			})

			fyne.Do(func() {
				progressBar.SetValue(1.0)
				fmt.Println("Process completed!")
				startButton.Enable()
			})
		}()
	})

	//--------------------------------------------------------------
	// Create UI labels and buttons
	//--------------------------------------------------------------
	actionLabel := widget.NewLabel("Action Type:")
	actionLabel.TextStyle.Bold = true

	languageLabel := widget.NewLabel("Language:")
	languageLabel.TextStyle.Bold = true

	fileLabel := widget.NewLabel("Audio File:")
	fileLabel.TextStyle.Bold = true

	// Output path label
	outputPathLabel := widget.NewLabel("Output Path:")
	outputPathLabel.TextStyle.Bold = true

	// Directory display labels
	directoryLabel = widget.NewLabel(fmt.Sprintf("Input Directory: %s", config.LastDirectory))
	directoryLabel.TextStyle.Italic = true

	outputDirectoryLabel = widget.NewLabel(fmt.Sprintf("Output Directory: %s", filepath.Dir(config.OutputPath)))
	outputDirectoryLabel.TextStyle.Italic = true

	progressLabel := widget.NewLabel("Progress:")
	progressLabel.TextStyle.Bold = true

	// Output field label
	outputLabel := widget.NewLabel("Output:")
	outputLabel.TextStyle.Bold = true
	// Clear output button
	clearOutputButton := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		outputField.SetText("")
		fmt.Println("Output cleared")
	})

	//--------------------------------------------------------------
	// Create left configuration panel
	//--------------------------------------------------------------
	leftPanel := container.NewVBox(
		widget.NewCard("Configuration", "Select your audio note processing options",
			container.NewVBox(
				actionLabel,
				container.NewHBox(
					actionSelect,
					refreshActionButton,
					newActionButton,
				),
				widget.NewSeparator(),
				languageLabel,
				languageSelect,
				widget.NewSeparator(),
				fileLabel,
				fileSelector,
				directoryLabel,
				widget.NewSeparator(),
				outputPathLabel,
				outputPathSelector,
				outputDirectoryLabel,
				widget.NewSeparator(),
				// Start button moved here, under directory line
				container.NewHBox(
					startButton,
					layout.NewSpacer(),
				),
			),
		),
		widget.NewSeparator(),
		container.NewVBox(
			progressLabel,
			progressBar,
			widget.NewSeparator(),
		),
		// Output label with clear button on the right
		container.NewBorder(
			// Top: Just the label
			container.NewPadded(outputLabel),
			// Bottom: Centered save button
			outputField,
			container.NewVBox(
				clearOutputButton,
			),
			// Left, Right:
			nil,

			nil,
			// Center: Maximized scrollable editor
		),
	)

	//--------------------------------------------------------------
	// Create main layout and set window content
	//--------------------------------------------------------------
	content := container.NewHSplit(leftPanel, rightPanel)

	// Add some padding around the content
	paddedContent := container.NewPadded(content)

	w.SetContent(paddedContent)

	//--------------------------------------------------------------
	// Set up window close handler and start application
	//--------------------------------------------------------------
	w.SetOnClosed(func() {
		// Restore original stdout
		if outputCapture != nil {
			outputCapture.Close()
		}
		config.Save()
	})

	w.ShowAndRun()
}
