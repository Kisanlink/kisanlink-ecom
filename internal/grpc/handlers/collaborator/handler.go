package collaborator

import (
	"github.com/Kisanlink/kisanlink-ecom/internal/aaa"
	"github.com/Kisanlink/kisanlink-ecom/internal/domain/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/saga"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/gst"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/otp"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler implements the CollaboratorService gRPC interface
type Handler struct {
	pb.UnimplementedCollaboratorServiceServer
	db           *gorm.DB
	logger       *logrus.Logger
	aaaClient    *aaa.Client
	gstService   *gst.Service
	sagaExecutor *saga.SagaExecutor
	otpService   otp.Service
	stateMachine *collaborator.StateMachine
}

// NewHandler creates a new collaborator handler
func NewHandler(
	db *gorm.DB,
	logger *logrus.Logger,
	aaaClient *aaa.Client,
	gstService *gst.Service,
	sagaExecutor *saga.SagaExecutor,
	otpService otp.Service,
	stateMachine *collaborator.StateMachine,
) *Handler {
	return &Handler{
		db:           db,
		logger:       logger,
		aaaClient:    aaaClient,
		gstService:   gstService,
		sagaExecutor: sagaExecutor,
		otpService:   otpService,
		stateMachine: stateMachine,
	}
}
