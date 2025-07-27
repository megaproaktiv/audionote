package translate

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	awsutil "github.com/megaproaktiv/audionote-config/aws"
)

var Client *transcribe.Client

func InitClient(ctx context.Context, profile string) error {
	cfg, err := awsutil.LoadAndValidateAWSConfig(ctx, profile)
	if err != nil {
		return err
	}

	Client = transcribe.NewFromConfig(cfg)
	return nil
}

// Translate converts an audio file and transcribes it using AWS Transcribe
// inputFile: path to the input audio file (M4A or MP3)
// bucket: S3 bucket name for storing temporary files
// languageCode: language code for transcription (e.g., "en-US", "de-DE")
func Translate(ctx context.Context, client *transcribe.Client, inputFile string, bucket string, languageCode string) string {

	mp3File, err := CopyFileToValidName(inputFile)
	if err != nil {
		log.Fatalf("Error copy file: %v", err)
	}

	mp3Key, err := CopyToS3(mp3File, bucket)
	if err != nil {
		log.Fatalf("Error copying file to S3: %v", err)
	}

	jobName, err := StartTranscribeJob(ctx, client, bucket, mp3Key, languageCode)
	if err != nil {
		log.Fatalf("Error starting transcription job: %v", err)
	}

	if err := WaitForTranscribeJob(jobName); err != nil {
		log.Fatalf("Error waiting for transcription job: %v", err)
	}

	transcript, err := GetTranscriptText(jobName, bucket)
	if err != nil {
		log.Fatalf("Error getting transcript text: %v", err)
	}

	// Clean up the copied file
	if err := os.Remove(mp3File); err != nil {
		log.Printf("Warning: Could not remove temporary file %s: %v", mp3File, err)
	} else {
		fmt.Printf("Cleaned up temporary file: %s", mp3File)
	}

	return transcript
}
