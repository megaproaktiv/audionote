package panel

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/megaproaktiv/audionote-config/configuration"
)

func SelectButton(panel *Panel, config *configuration.Config) (*dialog.FileDialog, error) {
	w := *panel.Window
	// change Directory to panel.CurrentDir
	os.Chdir(panel.CurrentDir)

	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		// Always restore original directory first
		configuration.RestoreDirectory(panel.CurrentDir)

		if err != nil {
			fmt.Printf("Error selecting output file: %v\n", err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		selectedPath := reader.URI().Path()
		config.OutputPath = selectedPath
		panel.OutputPathSelector.SetText(fmt.Sprintf("Selected: %s", filepath.Base(selectedPath)))
		fmt.Printf("Output path selected: %s\n", selectedPath)

		// Update output directory label
		panel.OutputDirectoryLabel.SetText(fmt.Sprintf("Output Directory: %s", filepath.Dir(selectedPath)))
		fmt.Printf("Updated output directory to: %s\n", filepath.Dir(selectedPath))
	}, w)

	// Set file filter for text files
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".md"}))

	// Set default filename
	dialog.SetFileName("result.txt")

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

	fmt.Printf("Opening output file dialog (should start in: %s)\n", config.LastDirectory)
	dialog.Show()

	return dialog, nil
}
