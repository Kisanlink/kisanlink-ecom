package collaborator

import (
	"kisanlink-ecom/internal/aaa"
	"kisanlink-ecom/internal/saga"
	"kisanlink-ecom/internal/services/gst"
	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"

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
}

// NewHandler creates a new collaborator handler
func NewHandler(
	db *gorm.DB,
	logger *logrus.Logger,
	aaaClient *aaa.Client,
	gstService *gst.Service,
	sagaExecutor *saga.SagaExecutor,
) *Handler {
	return &Handler{
		db:           db,
		logger:       logger,
		aaaClient:    aaaClient,
		gstService:   gstService,
		sagaExecutor: sagaExecutor,
	}
}
