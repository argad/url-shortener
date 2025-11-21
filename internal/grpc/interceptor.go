package grpc

import (
	"context"

	"github.com/argad/url-shortener/internal/auth"
	pb "github.com/argad/url-shortener/internal/grpc/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	userIDKey           = "user_id"
)

// AuthInterceptor provides authentication functionality for gRPC
type AuthInterceptor struct {
	logger *zap.Logger
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(logger *zap.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		logger: logger,
	}
}

// Unary returns a unary server interceptor for authentication
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			a.logger.Debug("No metadata in context")
			// Generate new user ID if no auth provided
			newUserID, err := auth.GenerateUserID()
			if err != nil {
				a.logger.Error("Failed to generate user ID", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to generate user ID")
			}
			// Inject into request if it has UserId field
			a.injectUserID(req, newUserID)
			return handler(ctx, req)
		}

		// Try to get authorization token
		tokens := md.Get(authorizationHeader)
		if len(tokens) == 0 {
			// No token, generate new user ID
			newUserID, err := auth.GenerateUserID()
			if err != nil {
				a.logger.Error("Failed to generate user ID", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to generate user ID")
			}
			a.injectUserID(req, newUserID)
			return handler(ctx, req)
		}

		// Validate token
		token := tokens[0]
		userID, err := auth.VerifyJWTToken(token)
		if err != nil {
			a.logger.Error("Invalid token", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Inject user ID into request
		a.injectUserID(req, userID)

		return handler(ctx, req)
	}
}

// injectUserID injects user ID into the request using reflection
func (a *AuthInterceptor) injectUserID(req interface{}, userID string) {
	// Set UserId field directly for all request types
	switch r := req.(type) {
	case *pb.ShortenURLRequest:
		r.UserId = userID
	case *pb.ShortenURLBatchRequest:
		r.UserId = userID
	case *pb.GetUserURLsRequest:
		r.UserId = userID
	case *pb.DeleteURLsRequest:
		r.UserId = userID
	}
}
