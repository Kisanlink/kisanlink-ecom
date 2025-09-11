package events

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient defines the interface for SQS operations
type SQSClient interface {
    SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}
