package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
)

// Config structure for storing application settings
type Config struct {
	LastActionType string `mapstructure:"last_action_type"`
	LastLanguage   string `mapstructure:"last_language"`
	LastDirectory  string `mapstructure:"last_directory"`
}

// loadPromptFiles reads all prompt-*.txt files from the config directory
func loadPromptFiles() ([]string, error) {
	var actionTypes []string
	
	err := filepath.WalkDir("./config", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasPrefix(d.Name(), "prompt-") && strings.HasSuffix(d.Name(), ".txt") {
			// Extract action type from filename: prompt-ACTION.txt -> ACTION
			actionType := strings.TrimPrefix(d.Name(), "prompt-")
			actionType = strings.TrimSuffix(actionType, ".txt")
			actionTypes = append(actionTypes, actionType)
		}
		
		return nil
	})
	
	return actionTypes, err
}

// initConfig initializes Viper configuration
func initConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	
	// Get user's Documents directory as default
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	documentsDir := filepath.Join(homeDir, "Documents")
	
	// Set defaults
	viper.SetDefault("last_action_type", "")
	viper.SetDefault("last_language", "en-US")
	viper.SetDefault("last_directory", documentsDir)
	
	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create it
			fmt.Println("Config file not found, creating new one...")
		} else {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
	}
	
	// Ensure last directory exists, fallback to Documents if not
	if config.LastDirectory == "" || !dirExists(config.LastDirectory) {
		config.LastDirectory = documentsDir
	}
	
	return &config
}

// saveConfig saves the current configuration
func saveConfig(config *Config) {
	viper.Set("last_action_type", config.LastActionType)
	viper.Set("last_language", config.LastLanguage)
	viper.Set("last_directory", config.LastDirectory)
	
	// Ensure config directory exists
	if err := os.MkdirAll("./config", 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}
	
	if err := viper.WriteConfigAs("./config/config.yaml"); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
	} else {
		fmt.Println("Configuration saved successfully")
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("Audio Note Configuration")
	w.Resize(fyne.NewSize(500, 400))

	// Initialize configuration
	config := initConfig()
	
	// Load prompt files to populate action types
	actionTypes, err := loadPromptFiles()
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
	if defaultAction == "" || !contains(actionTypes, defaultAction) {
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
		
		// Change to the last used directory if it exists
		if config.LastDirectory != "" && dirExists(config.LastDirectory) {
			os.Chdir(config.LastDirectory)
			fmt.Printf("Changed to directory: %s\n", config.LastDirectory)
		}
		
		dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Restore original directory
			os.Chdir(currentDir)
			
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
		
		fmt.Printf("Opening file dialog in directory: %s\n", config.LastDirectory)
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
		saveConfig(config)
		
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
		saveConfig(config)
		w.Close()
	})
	
	w.ShowAndRun()
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to check if directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
