package grpc

import (
	"context"
	"errors"

	pb "github.com/argad/url-shortener/internal/grpc/pb"
	"github.com/argad/url-shortener/internal/service"
	"github.com/argad/url-shortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the gRPC ShortenerService interface
type Server struct {
	pb.UnimplementedShortenerServiceServer
	urlService *service.URLService
	logger     *zap.Logger
}

// NewServer creates a new gRPC server instance
func NewServer(urlService *service.URLService, logger *zap.Logger) *Server {
	return &Server{
		urlService: urlService,
		logger:     logger,
	}
}

// ShortenURL implements the ShortenURL RPC method
func (s *Server) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID is required")
	}

	result, err := s.urlService.ShortenURL(req.UserId, req.Url)
	if err != nil {
		s.logger.Error("Failed to shorten URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	return &pb.ShortenURLResponse{
		ShortUrl:      result.ShortURL,
		AlreadyExists: result.AlreadyExists,
	}, nil
}

// ShortenURLBatch implements the ShortenURLBatch RPC method
func (s *Server) ShortenURLBatch(ctx context.Context, req *pb.ShortenURLBatchRequest) (*pb.ShortenURLBatchResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID is required")
	}
	if len(req.Items) == 0 {
		return &pb.ShortenURLBatchResponse{Items: []*pb.BatchResponseItem{}}, nil
	}

	// Convert protobuf items to service items
	items := make([]service.BatchItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.BatchItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		}
	}

	results, err := s.urlService.ShortenURLBatch(req.UserId, items)
	if err != nil {
		s.logger.Error("Failed to shorten batch", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to shorten batch")
	}

	// Convert service results to protobuf
	pbResults := make([]*pb.BatchResponseItem, len(results))
	for i, result := range results {
		pbResults[i] = &pb.BatchResponseItem{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
		}
	}

	return &pb.ShortenURLBatchResponse{Items: pbResults}, nil
}

// GetOriginalURL implements the GetOriginalURL RPC method
func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	if req.ShortCode == "" {
		return nil, status.Error(codes.InvalidArgument, "Short code is required")
	}

	originalURL, isDeleted, err := s.urlService.GetOriginalURL(req.ShortCode)
	if err != nil {
		// Check if URL was deleted
		var deletedErr *storage.URLDeletedError
		if errors.As(err, &deletedErr) {
			return &pb.GetOriginalURLResponse{
				OriginalUrl: "",
				IsDeleted:   true,
			}, nil
		}

		s.logger.Error("Failed to get original URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get original URL")
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl: originalURL,
		IsDeleted:   isDeleted,
	}, nil
}

// GetUserURLs implements the GetUserURLs RPC method
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID is required")
	}

	urls, err := s.urlService.GetUserURLs(req.UserId)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	// Convert service results to protobuf
	pbURLs := make([]*pb.UserURL, len(urls))
	for i, url := range urls {
		pbURLs[i] = &pb.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return &pb.GetUserURLsResponse{Urls: pbURLs}, nil
}

// DeleteURLs implements the DeleteURLs RPC method
func (s *Server) DeleteURLs(ctx context.Context, req *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID is required")
	}
	if len(req.ShortCodes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Short codes are required")
	}

	err := s.urlService.DeleteURLs(req.UserId, req.ShortCodes)
	if err != nil {
		s.logger.Error("Failed to delete URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete URLs")
	}

	return &pb.DeleteURLsResponse{Success: true}, nil
}

// GetStats implements the GetStats RPC method
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// Note: Network-based access control should be handled by infrastructure (firewall, proxy)
	// or via a separate interceptor checking client IP from context

	urlCount, userCount, err := s.urlService.GetStats()
	if err != nil {
		s.logger.Error("Failed to get stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get statistics")
	}

	return &pb.GetStatsResponse{
		Urls:  int32(urlCount),
		Users: int32(userCount),
	}, nil
}
