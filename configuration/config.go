package configuration

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/spf13/viper"
)

// Config structure for storing application settings
type Config struct {
	LastActionType string `mapstructure:"last_action_type"`
	LastLanguage   string `mapstructure:"last_language"`
	LastDirectory  string `mapstructure:"last_directory"`
	S3Bucket       string `mapstructure:"s3_bucket"`
	AWSProfile     string `mapstructure:"aws_profile"`
	Model          string `mapstructure:"model"`
	OutputLines    int    `mapstructure:"output_lines"`
	OutputPath     string `mapstructure:"output_path"`
}

var ConfigPath string

// LoadPromptFiles reads all prompt-*.txt files from the config directory
func LoadPromptFiles() ([]string, error) {
	var actionTypes []string

	err := filepath.WalkDir("./config", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasPrefix(d.Name(), "prompt-") && strings.HasSuffix(d.Name(), ".txt") {
			// Extract action type from filename: prompt-ACTION.txt -> ACTION
			actionType := strings.TrimPrefix(d.Name(), "prompt-")
			actionType = strings.TrimSuffix(actionType, ".txt")
			actionTypes = append(actionTypes, actionType)
		}

		return nil
	})

	return actionTypes, err
}

// InitConfigWithFS initializes Viper configuration with embedded filesystem and returns a Config struct
func InitConfigWithFS(defaultConfigFS fs.FS) *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	ConfigPath = path.Join(homeDir, ".config", "audionote")
	viper.AddConfigPath(ConfigPath)
	fmt.Println("Config path:", ConfigPath)
	// Get user's Documents directory as default
	documentsDir := filepath.Join(homeDir, "Documents")

	// Set defaults
	viper.SetDefault("last_action_type", "")
	viper.SetDefault("last_language", "en-US")
	viper.SetDefault("last_directory", documentsDir)
	viper.SetDefault("s3_bucket", "")
	viper.SetDefault("aws_profile", "default")
	viper.SetDefault("model", "anthropic.claude-3-5-sonnet-20240620-v1:0")
	viper.SetDefault("output_lines", 10)
	viper.SetDefault("output_path", filepath.Join(documentsDir, "result.txt"))

	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create it
			fmt.Println("Config file not found, creating new one...")

			CreateConfigFile(defaultConfigFS)

		} else {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
	}

	// Ensure last directory exists, fallback to Documents if not
	if config.LastDirectory == "" || !DirExists(config.LastDirectory) {
		config.LastDirectory = documentsDir
	}

	// Validate OutputLines - ensure it's between 5 and 50
	if config.OutputLines < 5 {
		config.OutputLines = 10
	} else if config.OutputLines > 50 {
		config.OutputLines = 50
	}

	return &config
}

// Save saves the current configuration to file
func (c *Config) Save() {
	viper.Set("last_action_type", c.LastActionType)
	viper.Set("last_language", c.LastLanguage)
	viper.Set("last_directory", c.LastDirectory)
	viper.Set("s3_bucket", c.S3Bucket)
	viper.Set("aws_profile", c.AWSProfile)
	viper.Set("model", c.Model)
	viper.Set("output_lines", c.OutputLines)
	viper.Set("output_path", c.OutputPath)

	if err := viper.WriteConfigAs(path.Join(ConfigPath, "config.yaml")); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
	} else {
		fmt.Println("Configuration saved successfully")
	}
}

// GetDirectoryURI returns a URI for the directory, with enhanced compatibility
func (c *Config) GetDirectoryURI() fyne.URI {
	if c.LastDirectory != "" && DirExists(c.LastDirectory) {
		// Ensure the path is absolute and clean
		absPath, err := filepath.Abs(c.LastDirectory)
		if err != nil {
			fmt.Printf("Error getting absolute path for %s: %v\n", c.LastDirectory, err)
			return storage.NewFileURI(c.LastDirectory)
		}
		return storage.NewFileURI(absPath)
	}
	return nil
}

// SetDirectoryForDialog attempts to set the starting directory for a file dialog
func (c *Config) SetDirectoryForDialog() error {
	if c.LastDirectory != "" && DirExists(c.LastDirectory) {
		// Change to the directory to ensure file dialog starts there
		err := os.Chdir(c.LastDirectory)
		if err != nil {
			fmt.Printf("Error changing to directory %s: %v\n", c.LastDirectory, err)
			return err
		}
		fmt.Printf("Successfully changed working directory to: %s\n", c.LastDirectory)
		return nil
	}
	return fmt.Errorf("directory does not exist: %s", c.LastDirectory)
}

// RestoreDirectory restores the original working directory
func RestoreDirectory(originalDir string) {
	if originalDir != "" {
		err := os.Chdir(originalDir)
		if err != nil {
			fmt.Printf("Error restoring directory to %s: %v\n", originalDir, err)
		} else {
			fmt.Printf("Restored working directory to: %s\n", originalDir)
		}
	}
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// TestDirectoryAccess tests if the stored directory is accessible
func (c *Config) TestDirectoryAccess() bool {
	if c.LastDirectory == "" {
		fmt.Println("No directory stored in config")
		return false
	}

	if !DirExists(c.LastDirectory) {
		fmt.Printf("Stored directory does not exist: %s\n", c.LastDirectory)
		return false
	}

	// Try to read the directory contents
	entries, err := os.ReadDir(c.LastDirectory)
	if err != nil {
		fmt.Printf("Cannot read directory %s: %v\n", c.LastDirectory, err)
		return false
	}

	fmt.Printf("Directory %s is accessible with %d entries\n", c.LastDirectory, len(entries))

	// Count audio files
	audioCount := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".mp3") || strings.HasSuffix(strings.ToLower(name), ".m4a") {
			audioCount++
			fmt.Printf("  Found audio file: %s\n", name)
		}
	}

	fmt.Printf("Found %d audio files in directory\n", audioCount)
	return true
}

// LoadPromptContent reads the content of a prompt file for the given action type
func LoadPromptContent(actionType string) (string, error) {
	filename := fmt.Sprintf("prompt-%s.txt", actionType)
	filepath := filepath.Join(ConfigPath, filename)
	fmt.Printf("Loading prompt file %s\n", filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %v", filename, err)
	}

	return string(content), nil
}

// SavePromptContent saves the content to a prompt file for the given action type
func SavePromptContent(actionType, content string) error {
	filename := fmt.Sprintf("prompt-%s.txt", actionType)
	filepath := filepath.Join(ConfigPath, filename)

	// Ensure config directory exists
	if err := os.MkdirAll("./config", 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write prompt file %s: %v", filename, err)
	}

	return nil
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// First time config
func CreateConfigFile(defaultConfigFS fs.FS) error {

	// Ensure config directory exists
	if err := os.MkdirAll(ConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	fmt.Printf("Created directory %v\n", ConfigPath)

	// Copy all files from embedded config-default directory
	err := fs.WalkDir(defaultConfigFS, "config-default", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "config-default" {
			return nil
		}

		// Get the relative path within config-default
		relPath, err := filepath.Rel("config-default", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Create the destination path
		destPath := filepath.Join(ConfigPath, relPath)

		if d.IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", destPath, err)
			}
			fmt.Printf("Created directory: %s\n", destPath)
		} else {
			// Copy file
			content, err := fs.ReadFile(defaultConfigFS, path)
			if err != nil {
				return fmt.Errorf("failed to read embedded file %s: %v", path, err)
			}

			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %v", destPath, err)
			}
			fmt.Printf("Copied file: %s\n", destPath)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy default config files: %v", err)
	}

	fmt.Println("Successfully copied all default configuration files")
	return nil
}
