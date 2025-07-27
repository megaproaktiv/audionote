package panel

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/megaproaktiv/audionote-config/configuration"
)

// updateOutputFieldSize updates the output field size based on configuration
func UpdateOutputFieldSize(p Panel, config *configuration.Config) {
	outputField := p.OutputField
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
