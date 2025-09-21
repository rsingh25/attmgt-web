FROM golang:1.24.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# RUN go tool sqlc generate .

RUN CGO_ENABLED=0 go build -o main main.go

FROM alpine:3.20.1 AS prod

# Create a new user and group
# RUN groupadd -r appgroup && useradd -r -g appgroup -s /bin/bash appuser

# Workgin directory
WORKDIR /app

COPY --from=build /app/main /app/main

#RUN chown -R appuser:appgroup /app

EXPOSE 80

# Switch to the non-root user
# USER appuser

CMD ["./main"]


