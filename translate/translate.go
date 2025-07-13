package translate

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/transcribe"

	"github.com/aws/aws-sdk-go-v2/config"
)

var Client *transcribe.Client

func InitClient(profile string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile))
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

	return transcript
}
