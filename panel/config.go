package panel

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsutil "github.com/megaproaktiv/audionote-config/aws"
	"github.com/megaproaktiv/audionote-config/configuration"
)

// validateS3Bucket checks if the S3 bucket exists and returns its region
func validateS3Bucket(bucketName, awsProfile string) (bool, string, string, error) {
	if bucketName == "" {
		return false, "", "Bucket name is empty", fmt.Errorf("bucket name is empty")
	}

	// Load AWS config using the common utility
	ctx := context.Background()
	cfg, err := awsutil.LoadAndValidateAWSConfig(ctx, awsProfile)
	if err != nil {
		return false, "", fmt.Sprintf("Failed to load/validate AWS config: %v", err), err
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Check if bucket exists by trying to get its location
	locationOutput, err := s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return false, "", fmt.Sprintf("Bucket '%s' does not exist", bucketName), err
		}
		if strings.Contains(err.Error(), "AccessDenied") {
			return false, "", fmt.Sprintf("Access denied to bucket '%s'. Check your AWS credentials and permissions.", bucketName), err
		}
		return false, "", fmt.Sprintf("Error checking bucket '%s': %v", bucketName, err), err
	}

	// Get bucket region
	bucketRegion := "us-east-1" // Default region for GetBucketLocation
	if locationOutput.LocationConstraint != "" {
		bucketRegion = string(locationOutput.LocationConstraint)
	}

	// Get current AWS config region
	currentRegion := cfg.Region

	// Check if regions match
	regionMatch := bucketRegion == currentRegion
	var regionMessage string
	if regionMatch {
		regionMessage = fmt.Sprintf("✓ Bucket region (%s) matches current AWS region", bucketRegion)
	} else {
		regionMessage = fmt.Sprintf("⚠ Bucket region (%s) differs from current AWS region (%s)", bucketRegion, currentRegion)
	}

	successMessage := fmt.Sprintf("✓ Bucket '%s' exists and is accessible\n%s", bucketName, regionMessage)
	return true, bucketRegion, successMessage, nil
}

// showConfigDialog displays the configuration dialog

func (p *Panel) ShowConfigDialog(config *configuration.Config) {

	outputField := p.OutputField
	w := p.Window

	// Create entry widgets for configuration
	s3BucketEntry := widget.NewEntry()
	s3BucketEntry.SetText(config.S3Bucket)
	s3BucketEntry.SetPlaceHolder("Enter S3 bucket name (e.g., my-audio-bucket)")

	awsProfileEntry := widget.NewEntry()
	awsProfileEntry.SetText(config.AWSProfile)
	awsProfileEntry.SetPlaceHolder("Enter AWS profile name (e.g., default)")

	// Create S3 bucket check button
	s3CheckButton := widget.NewButtonWithIcon("Check", theme.ConfirmIcon(), nil)
	s3CheckButton.OnTapped = func() {
		bucketName := strings.TrimSpace(s3BucketEntry.Text)
		awsProfile := strings.TrimSpace(awsProfileEntry.Text)
		if awsProfile == "" {
			awsProfile = "default"
		}

		// Change button text to indicate checking
		s3CheckButton.SetText("Checking...")
		s3CheckButton.Disable()

		// Perform validation in a goroutine to avoid blocking UI
		go func() {
			exists, region, message, err := validateS3Bucket(bucketName, awsProfile)

			// Update UI on main thread
			s3CheckButton.SetText("Check")
			s3CheckButton.Enable()

			var dialogTitle string
			var dialogIcon fyne.Resource

			if exists {
				dialogTitle = "S3 Bucket Validation - Success"
				dialogIcon = theme.ConfirmIcon()
			} else {
				dialogTitle = "S3 Bucket Validation - Error"
				dialogIcon = theme.ErrorIcon()
			}

			// Show result dialog
			resultDialog := dialog.NewCustom(
				dialogTitle,
				"OK",
				container.NewVBox(
					container.NewHBox(
						widget.NewIcon(dialogIcon),
						widget.NewLabel(dialogTitle),
					),
					widget.NewSeparator(),
					widget.NewRichTextFromMarkdown(fmt.Sprintf("**Bucket:** %s\n**AWS Profile:** %s\n\n%s", bucketName, awsProfile, message)),
				),
				*w,
			)
			resultDialog.Resize(fyne.NewSize(400, 200))
			resultDialog.Show()

			if err != nil {
				fmt.Printf("S3 bucket validation error: %v\n", err)
			} else {
				fmt.Printf("S3 bucket validation success: %s in region %s\n", bucketName, region)
			}
		}()
	}

	// Create S3 bucket container with entry and check button
	s3BucketContainer := container.NewBorder(nil, nil, nil, s3CheckButton, s3BucketEntry)

	// Create model entry
	modelEntry := widget.NewEntry()
	modelEntry.SetText(config.Model)
	modelEntry.SetPlaceHolder("Enter Bedrock model ID (e.g., anthropic.claude-3-5-sonnet-20240620-v1:0)")

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
	modelLabel := widget.NewRichTextFromMarkdown("**Bedrock Model:**\nThe AWS Bedrock model ID to use for processing (e.g., anthropic.claude-3-5-sonnet-20240620-v1:0).")
	outputPathLabel := widget.NewRichTextFromMarkdown("**Output File Path:**\nThe path where the processing result will be saved.")
	outputLabel := widget.NewRichTextFromMarkdown("**Output Display Lines:**\nMinimum number of lines to display in the output area (5-50).")

	// Create form content
	formContent := container.NewVBox(
		s3Label,
		s3BucketContainer,
		widget.NewSeparator(),
		awsLabel,
		awsProfileEntry,
		widget.NewSeparator(),
		modelLabel,
		modelEntry,
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
				model := modelEntry.Text
				outputPath := outputPathEntry.Text
				outputLines := int(outputLinesSlider.Value)

				if awsProfile == "" {
					awsProfile = "default"
				}

				if model == "" {
					model = "anthropic.claude-3-5-sonnet-20240620-v1:0"
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
				config.Model = model
				config.OutputPath = outputPath
				config.OutputLines = outputLines

				// Save configuration
				config.Save()

				// Update output field size if it changed
				if outputField != nil {
					p.OutputField = outputField
					p.UpdateOutputFieldSize(config)
				}

				// Show success message
				successMsg := fmt.Sprintf("Configuration saved successfully!\n\nS3 Bucket: %s\nAWS Profile: %s\nModel: %s\nOutput Path: %s\nOutput Lines: %d",
					s3Bucket, awsProfile, model, outputPath, outputLines)
				dialog.ShowInformation("Configuration Saved", successMsg, *w)

				fmt.Printf("Configuration updated - S3 Bucket: %s, AWS Profile: %s, Model: %s, Output Path: %s, Output Lines: %d\n",
					config.S3Bucket, config.AWSProfile, config.Model, config.OutputPath, config.OutputLines)
			}
		},
		*w,
	)

	configDialog.Resize(fyne.NewSize(500, 500))
	configDialog.Show()
}
