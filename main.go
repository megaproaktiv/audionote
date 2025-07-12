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
	w := a.NewWindow("Audio Note Configuration")
	w.Resize(fyne.NewSize(500, 400))

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

	// Create the select widgets
	actionSelect := widget.NewSelect(
		actionTypes,
		func(value string) {
			fmt.Printf("Action selected: %s\n", value)
			config.LastActionType = value
		},
	)
	
	// Set default selection from config or first available option
	defaultAction := config.LastActionType
	if defaultAction == "" || !configuration.Contains(actionTypes, defaultAction) {
		defaultAction = actionTypes[0]
	}
	actionSelect.SetSelected(defaultAction)
	config.LastActionType = defaultAction

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

	// Create the main content with a nice layout
	content := container.NewVBox(
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
