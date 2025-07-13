package translate

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Call ffmpeg
// Transscript can not read m4a
func ConvertM4AToMP3(inputFile string) (string, error) {
	// Member must satisfy regular expression pattern: ^[0-9a-zA-Z._-]+
	validInputFile, err := CopyFileToValidName(inputFile)
	if err != nil {
		return "", err
	}
	outputFile := strings.TrimSuffix(validInputFile, ".m4a") + ".mp3"
	fmt.Printf("Converting %s to %s...\n", validInputFile, outputFile)
	cmd := exec.Command("ffmpeg", "-i", inputFile, outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return outputFile, nil
}

// CopyFileToValidName copies the file at src to a new valid name in the same directory.
// It returns the new file name, or an error.
func CopyFileToValidName(src string) (string, error) {
	re := regexp.MustCompile(`[0-9a-zA-Z._-]+`)
	dir := filepath.Dir(src)
	base := filepath.Base(src)
	ext := filepath.Ext(base)

	// Get base name without extension, sanitize
	namePart := base[:len(base)-len(ext)]
	sanitizedBase := ""
	for _, m := range re.FindAllString(namePart, -1) {
		sanitizedBase += m
	}
	if sanitizedBase == "" {
		sanitizedBase = "copy"
	}
	newName := sanitizedBase + ext
	dst := filepath.Join(dir, newName)

	// If src == dst, append _copy before ext
	if dst == src {
		newName = sanitizedBase + "_copy" + ext
		dst = filepath.Join(dir, newName)
	}

	// Copy file
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}

	return newName, nil
}
