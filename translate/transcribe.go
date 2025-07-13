package translate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TranscriptResponse struct {
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
	} `json:"results"`
}

func StartTranscribeJob(bucket, mp3Key string) (string, error) {
	jobName := strings.TrimSuffix(filepath.Base(mp3Key), ".mp3") + "-DMIN-" + fmt.Sprintf("%d", time.Now().Unix())
	mediaURI := fmt.Sprintf("s3://%s/%s", bucket, mp3Key)
	fmt.Printf("Starting transcription job '%s' for %s...\n", jobName, mediaURI)
	outputKey := fmt.Sprintf("summary/output/%s.json", jobName)
	cmd := exec.Command("aws", "transcribe", "start-transcription-job",
		"--transcription-job-name", jobName,
		"--language-code", "de-DE",
		"--media-sample-rate-hertz", "48000",
		"--media-format", "mp3",
		"--media", fmt.Sprintf("MediaFileUri=%s", mediaURI),
		"--output-bucket-name", bucket,
		"--output-key", outputKey,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return jobName, nil
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
