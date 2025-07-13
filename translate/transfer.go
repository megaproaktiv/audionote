package translate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CopyToS3(file, bucket string) (string, error) {
	s3Key := "summary/" + filepath.Base(file)
	dest := fmt.Sprintf("s3://%s/%s", bucket, s3Key)
	fmt.Printf("Copying %s to %s...\n", file, dest)
	cmd := exec.Command("aws", "s3", "cp", file, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return s3Key, nil
}
