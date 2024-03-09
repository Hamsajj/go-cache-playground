FROM golang:1.22-alpine as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 go build -v -o server.out


FROM alpine

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server.out /opt/app/

# Run the web service on container startup.
CMD ["/opt/app/server.out"]
