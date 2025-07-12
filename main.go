package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/megaproaktiv/audionote-config/configuration"
)

func main() {
	a := app.New()
	w := a.NewWindow("Audio Note LLM")
	w.Resize(fyne.NewSize(1000, 600)) // Increased width for editor

	// Initialize configuration
	config := configuration.InitConfig()
	
	// Load prompt files to populate action types
	actionTypes, err := configuration.LoadPromptFiles()
	if err != nil {
		fmt.Printf("Error loading prompt files: %v\n", err)
		actionTypes = []string{"summary", "call to action", "criticize"} // fallback
	}
	
	if len(actionTypes) == 0 {
		actionTypes = []string{"summary", "call to action", "criticize"} // fallback
	}
	
	fmt.Printf("Loaded action types: %v\n", actionTypes)

	// Create editor field for prompt content
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

	// Create the select widgets
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
	
	// Load initial prompt content
	loadPromptContent(defaultAction)

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

	// Create file selector for audio files
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

	// Create progress bar
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0.0)

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

	// Create the start button
	var startButton *widget.Button
	startButton = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		action := actionSelect.Selected
		language := languageSelect.Selected
		
		if selectedFilePath == "" {
			dialog.ShowInformation("No File Selected", "Please select an audio file first.", w)
			return
		}
		
		// Save current configuration
		config.Save()
		
		fmt.Printf("Starting process with Action: %s, Language: %s, File: %s\n", action, language, selectedFilePath)
		
		// Simulate progress
		go func() {
			startButton.Disable()
			for i := 0; i <= 100; i += 10 {
				progressBar.SetValue(float64(i) / 100.0)
				time.Sleep(200 * time.Millisecond)
			}
			progressBar.SetValue(1.0)
			fmt.Println("Process completed!")
			startButton.Enable()
		}()
	})

	// Create form layout with labels
	actionLabel := widget.NewLabel("Action Type:")
	actionLabel.TextStyle.Bold = true
	
	languageLabel := widget.NewLabel("Language:")
	languageLabel.TextStyle.Bold = true
	
	fileLabel := widget.NewLabel("Audio File:")
	fileLabel.TextStyle.Bold = true
	
	// Directory display label
	directoryLabel = widget.NewLabel(fmt.Sprintf("Directory: %s", config.LastDirectory))
	directoryLabel.TextStyle.Italic = true
	
	progressLabel := widget.NewLabel("Progress:")
	progressLabel.TextStyle.Bold = true

	// Prompt editor label
	promptLabel := widget.NewLabel("Prompt Editor:")
	promptLabel.TextStyle.Bold = true

	// Create the left side configuration panel
	leftPanel := container.NewVBox(
		widget.NewCard("Configuration", "Select your audio note processing options", 
			container.NewVBox(
				actionLabel,
				actionSelect,
				widget.NewSeparator(),
				languageLabel,
				languageSelect,
				widget.NewSeparator(),
				fileLabel,
				fileSelector,
				directoryLabel,
			),
		),
		widget.NewSeparator(),
		container.NewVBox(
			progressLabel,
			progressBar,
		),
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			startButton,
			layout.NewSpacer(),
		),
	)

	// Create the right side editor panel - fully maximized editor
	rightPanel := container.NewBorder(
		// Top: Just the label
		container.NewPadded(promptLabel),
		// Bottom: Centered save button
		container.NewPadded(
			container.NewHBox(
				layout.NewSpacer(),
				savePromptButton,
				layout.NewSpacer(),
			),
		),
		// Left, Right: nil
		nil, nil,
		// Center: Maximized scrollable editor
		container.NewScroll(promptEditor),
	)

	// Create horizontal split with left and right panels
	content := container.NewHSplit(leftPanel, rightPanel)
	content.SetOffset(0.5) // Equal split

	// Add some padding around the content
	paddedContent := container.NewPadded(content)

	w.SetContent(paddedContent)
	
	// Save config when window closes
	w.SetCloseIntercept(func() {
		config.Save()
		w.Close()
	})
	
	w.ShowAndRun()
}
