package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
)

type TranscriptResponse struct {
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
	} `json:"results"`
}

// StartTranscribeJob starts an AWS Transcribe job with the specified language code
// Supported language codes include: en-US, de-DE, fr-FR, es-ES, etc.
// See AWS Transcribe documentation for full list of supported languages
func StartTranscribeJob(ctx context.Context, client *transcribe.Client, bucket, mp3Key, languageCode string) (string, error) {
	jobName := strings.TrimSuffix(filepath.Base(mp3Key), ".mp3") + "-DMIN-" + fmt.Sprintf("%d", time.Now().Unix())
	mediaURI := fmt.Sprintf("s3://%s/%s", bucket, mp3Key)
	fmt.Printf("Starting transcription job '%s' for %s with language %s...\n", jobName, mediaURI, languageCode)
	outputKey := fmt.Sprintf("summary/output/%s.json", jobName)
	mediaFormat := types.MediaFormatM4a

	// Convert language code to AWS Transcribe format
	var languageCodeType types.LanguageCode
	switch languageCode {
	case "en-US":
		languageCodeType = types.LanguageCodeEnUs
	case "de-DE":
		languageCodeType = types.LanguageCodeDeDe
	default:
		languageCodeType = types.LanguageCodeEnUs
	}

	params := transcribe.StartTranscriptionJobInput{
		Media:                &types.Media{MediaFileUri: &mediaURI},
		TranscriptionJobName: &jobName,
		LanguageCode:         languageCodeType,
		MediaFormat:          mediaFormat,
		MediaSampleRateHertz: aws.Int32(48000),
		OutputBucketName:     &bucket,
		OutputKey:            &outputKey,
	}
	resp, err := client.StartTranscriptionJob(ctx, &params)
	if err != nil {
		return "", err
	}
	return *resp.TranscriptionJob.TranscriptionJobName, nil
}

func WaitForTranscribeJob(jobName string) error {
	fmt.Printf("Waiting for transcription job '%s' to complete...\n", jobName)
	for {
		cmd := exec.Command("aws", "transcribe", "get-transcription-job",
			"--transcription-job-name", jobName)
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		var resp map[string]any
		err = json.Unmarshal(output, &resp)
		if err != nil {
			return err
		}
		job, ok := resp["TranscriptionJob"].(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response format")
		}
		status, ok := job["TranscriptionJobStatus"].(string)
		if !ok {
			return fmt.Errorf("unexpected status format")
		}
		fmt.Printf("Current status: %s\n", status)
		if status == "COMPLETED" {
			break
		} else if status == "FAILED" {
			return fmt.Errorf("transcription job failed")
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}

func GetTranscriptText(jobName, bucket string) (string, error) {
	s3Key := fmt.Sprintf("summary/output/%s.json", jobName)
	localFile := s3Key
	if err := os.MkdirAll(filepath.Dir(localFile), os.ModePerm); err != nil {
		return "", err
	}
	s3Path := fmt.Sprintf("s3://%s/%s", bucket, s3Key)
	fmt.Printf("Fetching transcription result from %s...\n", s3Path)
	cmd := exec.Command("aws", "s3", "cp", s3Path, localFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(localFile)
	if err != nil {
		return "", err
	}
	var transcriptResp TranscriptResponse
	if err := json.Unmarshal(data, &transcriptResp); err != nil {
		return "", err
	}
	if len(transcriptResp.Results.Transcripts) == 0 {
		return "", fmt.Errorf("no transcript found")
	}
	return transcriptResp.Results.Transcripts[0].Transcript, nil
}
