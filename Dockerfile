FROM golang:1.25.1-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

# Copy dependency project
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/main.go

FROM alpine:3.22

WORKDIR /app
RUN apk add --no-cache make

# Copy the built binary
COPY --from=builder /build/server ./server

RUN chmod +x server

EXPOSE 3000

CMD [ "./server"]