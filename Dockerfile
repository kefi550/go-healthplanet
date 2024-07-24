FROM golang:1.22-bookworm as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd ./cmd

WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /healthplanet

CMD ["/healthplanet"]
