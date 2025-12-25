package server

import (
	"aquascore/internal/db/mongo"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Server holds dependencies for an HTTP server.
type Server struct {
	router     *gin.Engine
	grpcClient GrpcClient
}

// NewHTTPServer creates a new Server instance, setting up API routes.
func NewHTTPServer(store *mongo.Stores, analysisServerAddr string) (*Server, error) {
	grpcClient, err := newGRPCClient(analysisServerAddr)
	if err != nil {
		return nil, err
	}
	// Initialize the server with a new Gin router
	s := &Server{
		router:     gin.Default(),
		grpcClient: grpcClient,
	}
	initAPIHandler(
		s.router.Group("/api/v1").Use(otelgin.Middleware("API Server")),
		store, grpcClient,
	)
	return s, nil
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) Close() error {
	return s.grpcClient.Close()
}
