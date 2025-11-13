package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/mock"

	"kisanlink-ecom/internal/services/events"
)

// MockSQSClient is a mock implementation of SQS client
type MockSQSClient struct {
	mock.Mock
}

// Ensure MockSQSClient implements the SQSClient interface
var _ events.SQSClient = (*MockSQSClient)(nil)

// SendMessage mocks the SendMessage method
func (m *MockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}
