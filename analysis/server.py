import grpc
from concurrent import futures
import time
import logging
import os

# OpenTelemetry imports
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.sdk.resources import SERVICE_NAME, Resource # SERVICE_NAME is also from sdk.resources

# Import generated classes
# from generated import analysis_pb2
# from generated import analysis_pb2_grpc
from analysis.v1 import analysis_pb2, analysis_pb2_grpc
from grpc_reflection.v1alpha import reflection

# Import analysis logic
from logic.overview import analyze_performance_overview
from logic.comparison import analyze_result_comparison

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

class AnalysisServicer(analysis_pb2_grpc.AnalysisServiceServicer):
    """
    Implements the gRPC AnalysisService.
    """

    def AnalyzePerformanceOverview(self, request, context):
        """
        Handles the RPC for analyzing an athlete's performance overview.
        """
        logging.info(f"Received AnalyzePerformanceOverview request for athlete: {request.athlete_name}")
        try:
            # Delegate the core logic to the specialized function
            analyses = analyze_performance_overview(request.results)
            logging.info(f"Successfully analyzed {len(analyses)} events for {request.athlete_name}.")
            return analysis_pb2.AnalyzePerformanceOverviewResponse(event_analyses=analyses)
        except Exception as e:
            logging.error(f"Error in AnalyzePerformanceOverview for {request.athlete_name}: {e}", exc_info=True)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"An internal error occurred: {e}")
            return analysis_pb2.AnalyzePerformanceOverviewResponse()

    def AnalyzeResultComparison(self, request, context):
        """
        Handles the RPC for comparing a race result against the competition and records.
        """
        logging.info(f"Received AnalyzeResultComparison request for athlete: {request.target_result.athlete_name}")
        try:
            # Delegate the core logic to the specialized function
            comparisons = analyze_result_comparison(
                request.target_result,
                request.competition_results,
                request.records
            )
            logging.info(f"Successfully generated {len(comparisons)} comparisons.")
            return analysis_pb2.AnalyzeResultComparisonResponse(results_comparison=comparisons)
        except Exception as e:
            logging.error(f"Error in AnalyzeResultComparison: {e}", exc_info=True)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"An internal error occurred: {e}")
            return analysis_pb2.AnalyzeResultComparisonResponse()

def serve():
    """
    Starts the gRPC server and listens for requests.
    """
    port = os.getenv("GRPC_PORT", "50051")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Register the servicer
    analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
    
    # Enable server reflection
    SERVICE_NAMES = (
        analysis_pb2.DESCRIPTOR.services_by_name['AnalysisService'].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)
    
    # Start the server
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logging.info(f"✅ gRPC server started successfully on port {port}.")
    
    # Keep the server running
    try:
        while True:
            time.sleep(3600)  # Sleep for one hour
    except KeyboardInterrupt:
        logging.info("Attempting graceful shutdown...")
        server.stop(0)
        logging.info("Server stopped.")

def configure_opentelemetry():
    # Resource: identifies your service in the tracing system
    resource = Resource.create({
        SERVICE_NAME: os.getenv("OTEL_SERVICE_NAME", "AquaScore-Analysis-GRPC"),
    })

    # Exporter: sends spans to the OTLP collector (Jaeger)
    # The endpoint should be configurable via environment variable
    otlp_exporter = OTLPSpanExporter(
        endpoint=os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://jaeger.tracing.orb.local:4318/v1/traces"),
        timeout=5,
    )

    # Processor: batches and exports spans
    span_processor = BatchSpanProcessor(otlp_exporter)

    # Provider: manages the lifecycle of traces
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(span_processor)

    # Set the global tracer provider
    trace.set_tracer_provider(provider)

    # Instrument gRPC
    grpc_instrumentor = GrpcInstrumentorServer()
    grpc_instrumentor.instrument()
    logging.info("gRPC instrumentation initialized.")

    # Instrument logging
    LoggingInstrumentor().instrument(set_logging_format=True)
    logging.info("Logging instrumentation initialized.")


if __name__ == '__main__':
    configure_opentelemetry()
    serve()
