FROM python:3.12-slim

WORKDIR /app

# 1. Install dependencies
# Copy requirements.txt from the build context root
COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt --extra-index-url https://buf.build/gen/python

# 2. Copy source code
# Pants preserves the directory structure.
COPY grpcanalysis grpcanalysis

# 3. Set Python path
# Ensure python can find the modules in /app
ENV PYTHONPATH="/app"

# 4. Run the server
# Run as a module. Since server.py is in grpcanalysis/server.py
CMD ["python", "-m", "grpcanalysis.server"]