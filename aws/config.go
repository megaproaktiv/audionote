package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// LoadAWSConfig loads AWS configuration with environment variable support
// It checks for AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_SESSION_TOKEN
// If environment variables are present, it uses them; otherwise falls back to profile
func LoadAWSConfig(ctx context.Context, profile string) (aws.Config, error) {
	var cfg aws.Config
	var err error

	// Check for AWS environment variables
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	if accessKeyID != "" && secretAccessKey != "" {
		// Use environment variables for credentials
		fmt.Printf("Using AWS credentials from environment variables\n")
		if sessionToken != "" {
			// All three environment variables are present
			fmt.Printf("Using temporary credentials with session token\n")
			cfg, err = config.LoadDefaultConfig(ctx,
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					accessKeyID, secretAccessKey, sessionToken)))
		} else {
			// Only access key and secret key are present
			fmt.Printf("Using long-term credentials (access key + secret key)\n")
			cfg, err = config.LoadDefaultConfig(ctx,
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					accessKeyID, secretAccessKey, "")))
		}
	} else {
		// Fall back to profile-based configuration
		fmt.Printf("Using AWS profile: %s\n", profile)
		if profile != "" && profile != "default" {
			cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		} else {
			cfg, err = config.LoadDefaultConfig(ctx)
		}
	}

	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return cfg, nil
}

// ValidateAWSConfig validates the AWS configuration by calling STS GetCallerIdentity
func ValidateAWSConfig(ctx context.Context, cfg aws.Config) error {
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to verify AWS identity: %w", err)
	}

	fmt.Printf("AWS identity validated - Account: %s, User/Role: %s\n", 
		aws.ToString(identity.Account), aws.ToString(identity.Arn))
	return nil
}

// LoadAndValidateAWSConfig is a convenience function that loads and validates AWS config
func LoadAndValidateAWSConfig(ctx context.Context, profile string) (aws.Config, error) {
	cfg, err := LoadAWSConfig(ctx, profile)
	if err != nil {
		return cfg, err
	}

	err = ValidateAWSConfig(ctx, cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
