package events

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// EventPublishingConfig holds configuration for event publishing
type EventPublishingConfig struct {
    // SQS Configuration
    AWSRegion   string `json:"aws_region" env:"AWS_REGION" default:"us-east-1"`
    SQSQueueURL string `json:"sqs_queue_url" env:"SQS_QUEUE_URL"`

    // Retry Configuration
    MaxRetryAttempts int `json:"max_retry_attempts" env:"MAX_RETRY_ATTEMPTS" default:"5"`

    // Processing Configuration
    BatchSize          int `json:"batch_size" env:"EVENT_BATCH_SIZE" default:"100"`
    ProcessingInterval int `json:"processing_interval_seconds" env:"PROCESSING_INTERVAL_SECONDS" default:"30"`

    // Cleanup Configuration
    CleanupIntervalDays int `json:"cleanup_interval_days" env:"CLEANUP_INTERVAL_DAYS" default:"7"`
}

// NewSQSClient creates a new SQS client with the provided configuration
func NewSQSClient(ctx context.Context, region string) (*sqs.Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    return sqs.NewFromConfig(cfg), nil
}

// ValidateConfig validates the event publishing configuration
func (c *EventPublishingConfig) ValidateConfig() error {
    if c.SQSQueueURL == "" {
        return fmt.Errorf("SQS queue URL is required")
    }

    if c.AWSRegion == "" {
        return fmt.Errorf("AWS region is required")
    }

    if c.MaxRetryAttempts < 1 {
        return fmt.Errorf("max retry attempts must be at least 1")
    }

    if c.BatchSize < 1 {
        return fmt.Errorf("batch size must be at least 1")
    }

    if c.ProcessingInterval < 1 {
        return fmt.Errorf("processing interval must be at least 1 second")
    }

    return nil
}
