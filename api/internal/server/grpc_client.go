package server

import (
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"buf.build/gen/go/aqua/analysis/grpc/go/analysis/v1/analysisv1grpc"
)

type GrpcClient interface {
	analysisv1grpc.AnalysisServiceClient
	Close() error
}

type grpcClient struct {
	analysisv1grpc.AnalysisServiceClient
	conn *grpc.ClientConn
}

func newGRPCClient(addr string) (GrpcClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("ANALYSIS_SERVICE_ADDR is not set")
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	client := analysisv1grpc.NewAnalysisServiceClient(conn)
	return &grpcClient{
		AnalysisServiceClient: client,
		conn:                  conn,
	}, nil
}

func (grpc *grpcClient) Close() error {
	return grpc.conn.Close()
}
