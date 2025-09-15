# -------- Base build image --------
FROM golang:1.22 AS build
WORKDIR /src

# Install tools and swag for Swagger documentation
RUN --mount=type=cache,target=/var/cache/apt <<EOF
apt-get update -y && apt-get install -y --no-install-recommends ca-certificates && update-ca-certificates
EOF

# Install swag tool for generating Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

# Use Go module cache for faster builds
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy the rest and build
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/api/main.go -o docs

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# -------- Runtime image --------
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/api /app/api
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
