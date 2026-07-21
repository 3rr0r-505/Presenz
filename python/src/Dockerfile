# -------------------------
# Base Image
# -------------------------
FROM debian:latest

# -------------------------
# Install dependencies
# -------------------------
RUN apt-get update && \
    apt-get install -y python3 python3-venv python3-pip curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# -------------------------
# Set working directory
# -------------------------
WORKDIR /app

# -------------------------
# Copy project files
# -------------------------
COPY . /app

# -------------------------
# Create virtual environment
# -------------------------
RUN python3 -m venv /opt/venv

# Activate venv and install requirements if any
RUN /opt/venv/bin/pip install --upgrade pip && /opt/venv/bin/pip install -r requirements.txt

# -------------------------
# Use venv for Python
# -------------------------
ENV PATH="/opt/venv/bin:$PATH"

# -------------------------
# Expose port for server
# -------------------------
EXPOSE 8080

# -------------------------
# Default entrypoint
# -------------------------
ENTRYPOINT ["python3", "main.py"]
