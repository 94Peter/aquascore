gen_proto:
	python3 -m grpc_tools.protoc \
    -Igenerated=./proto \
    --python_out=./analysis \
    --grpc_python_out=./analysis \
    ./proto/analysis.proto
	mkdir -p ./internal/generated/analysis && \
	protoc --proto_path=./proto \
       --go_out=./internal/generated/analysis \
       --go_opt=paths=source_relative \
       --go-grpc_out=./internal/generated/analysis \
       --go-grpc_opt=paths=source_relative \
       ./proto/analysis.proto


run_analysis_server:
    pip install -r analysis/requirements.txt