package main

import (
	"fmt"
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
)

func main() {
	a := app.New()
	w := a.NewWindow("Audio Note Configuration")
	w.Resize(fyne.NewSize(450, 350))

	// Create the select widgets
	actionSelect := widget.NewSelect(
		[]string{"summary", "call to action", "criticize"},
		func(value string) {
			fmt.Printf("Action selected: %s\n", value)
		},
	)
	actionSelect.SetSelected("summary") // Set default selection

	languageSelect := widget.NewSelect(
		[]string{"en-US", "de-DE"},
		func(value string) {
			fmt.Printf("Language selected: %s\n", value)
		},
	)
	languageSelect.SetSelected("en-US") // Set default selection

	// Create file selector for audio files
	var selectedFilePath string
	var fileSelector *widget.Button
	fileSelector = widget.NewButton("Select Audio File", func() {
		dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
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
		}, w)
		
		// Set file filter for audio files
		dialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3", ".wav", ".m4a", ".aac", ".ogg", ".flac"}))
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
	w.ShowAndRun()
}
