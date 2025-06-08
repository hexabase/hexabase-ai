# Stage 1: Build stage to install dependencies
FROM python:3.11-slim as builder

WORKDIR /usr/src/app

# Install build tools
RUN pip install --no-cache-dir --upgrade pip

# Copy only the requirements file to leverage Docker cache
COPY requirements.txt ./

# Install dependencies
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# Stage 2: Final stage
FROM python:3.11-slim

WORKDIR /usr/src/app

# Copy the installed dependencies from the builder stage
COPY --from=builder /install /usr/local

# Copy the application code
COPY ./app ./app

# Expose the port the app runs on
EXPOSE 8000

# Command to run the application
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]