FROM python:3.12-slim

WORKDIR /app

# 1. Install dependencies
# This ensures that platform-specific wheels (numpy, grpcio) are installed correctly
# for the target architecture (linux/amd64 or linux/arm64)
COPY analysis/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt --extra-index-url https://buf.build/gen/python

# 2. Copy source code
# We copy the contents of the analysis directory directly into /app
# This prevents the local 'analysis' folder from shadowing the installed 'analysis' package
COPY analysis /app

# 3. Set Python path
# Ensure python can find the modules in /app (like logic)
ENV PYTHONPATH="/app"

# 4. Run the server
# Run as a top-level module
CMD ["python", "-m", "server"]