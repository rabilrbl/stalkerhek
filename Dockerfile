FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy source files from the host computer to the container
COPY . .

# Build the Go app with optimizations
RUN go build -ldflags="-s -w" -trimpath -o /app/stalkerhek cmd/stalkerhek/main.go

# Stage 2: Create the final minimal image
# skipcq: DOK-DL3007
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go executable from the previous stage
COPY --from=builder /app/stalkerhek .

# Expose port 5001 to the outside world
EXPOSE 8080

ENV PORT=8080

# Command to run the executable with arguments
# The CMD instruction has been replaced with ENTRYPOINT to allow arguments
ENTRYPOINT ["./stalkerhek"]
