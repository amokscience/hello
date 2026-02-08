package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func getAWSSecret(secretName string) (map[string]interface{}, error) {
	// Get AWS credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	// If no credentials in env, return empty (graceful fallback)
	if accessKey == "" || secretKey == "" {
		logger.Warn("AWS credentials not configured, skipping secrets fetch")
		return make(map[string]interface{}), nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		return make(map[string]interface{}), nil
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		logger.Warn("Failed to get secret", "error", err)
		return make(map[string]interface{}), nil
	}

	// Parse the secret as JSON
	var secretData map[string]interface{}
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		logger.Warn("Failed to parse secret JSON", "error", err)
		return make(map[string]interface{}), nil
	}

	return secretData, nil
}
