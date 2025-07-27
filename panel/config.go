package panel

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/megaproaktiv/audionote-config/configuration"
)

// showConfigDialog displays the configuration dialog

func ShowConfigDialog(config *configuration.Config, myPanel Panel) {

	outputField := myPanel.OutputField
	w := myPanel.Window

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
					parms := Panel{
						OutputField: outputField,
					}
					UpdateOutputFieldSize(parms, config)
				}

				// Show success message
				successMsg := fmt.Sprintf("Configuration saved successfully!\n\nS3 Bucket: %s\nAWS Profile: %s\nOutput Path: %s\nOutput Lines: %d",
					s3Bucket, awsProfile, outputPath, outputLines)
				dialog.ShowInformation("Configuration Saved", successMsg, *w)

				fmt.Printf("Configuration updated - S3 Bucket: %s, AWS Profile: %s, Output Path: %s, Output Lines: %d\n",
					config.S3Bucket, config.AWSProfile, config.OutputPath, config.OutputLines)
			}
		},
		*w,
	)

	configDialog.Resize(fyne.NewSize(500, 450))
	configDialog.Show()
}
