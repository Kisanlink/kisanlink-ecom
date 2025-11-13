package interceptors

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
	UserID    string                 `json:"user_id,omitempty"`
	FPOID     string                 `json:"fpo_id,omitempty"`
	Method    string                 `json:"method"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuditUnaryInterceptor logs audit trail for mutations
func AuditUnaryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only audit mutations (Create, Update, Delete, Deactivate)
		if !isMutation(info.FullMethod) {
			return handler(ctx, req)
		}

		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Build audit log
		auditLog := buildAuditLog(ctx, info.FullMethod, start, err)

		// Log audit event
		logEntry := logger.WithFields(logrus.Fields{
			"audit":      true,
			"timestamp":  auditLog.Timestamp,
			"request_id": auditLog.RequestID,
			"user_id":    auditLog.UserID,
			"fpo_id":     auditLog.FPOID,
			"method":     auditLog.Method,
			"status":     auditLog.Status,
			"duration":   auditLog.Duration,
			"ip_address": auditLog.IPAddress,
		})

		if err != nil {
			logEntry.WithError(err).Warn("Audit: mutation failed")
		} else {
			logEntry.Info("Audit: mutation succeeded")
		}

		// TODO: Store in database or audit log system for compliance
		// saveAuditLog(ctx, auditLog)

		return resp, err
	}
}

func buildAuditLog(ctx context.Context, method string, start time.Time, err error) AuditLog {
	auditLog := AuditLog{
		Timestamp: time.Now(),
		RequestID: GetRequestID(ctx),
		Method:    method,
		Duration:  time.Since(start),
	}

	// Get user information from claims
	if claims := GetUserClaims(ctx); claims != nil {
		if userID, ok := claims["user_id"].(string); ok {
			auditLog.UserID = userID
		}
		if fpoID, ok := claims["fpo_id"].(string); ok {
			auditLog.FPOID = fpoID
		}
	}

	// Get IP address and user agent from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			auditLog.IPAddress = ips[0]
		}
		if agents := md.Get("user-agent"); len(agents) > 0 {
			auditLog.UserAgent = agents[0]
		}
	}

	// Set status
	if err != nil {
		auditLog.Status = status.Code(err).String()
		auditLog.Error = err.Error()
	} else {
		auditLog.Status = codes.OK.String()
	}

	return auditLog
}

func isMutation(method string) bool {
	mutations := []string{
		"Create",
		"Update",
		"Delete",
		"Deactivate",
		"Verify",
	}

	for _, mutation := range mutations {
		if contains(method, mutation) {
			return true
		}
	}
	return false
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str[len(str)-len(substr):] == substr ||
		len(str) > len(substr) && str[len(str)-len(substr)-1:len(str)-len(substr)] == "/" && str[len(str)-len(substr):] == substr)
}
