FROM golang:1.22

RUN mkdir -p /app/cmd
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY test.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /test-runner
# RUN chmod +x /app

CMD ["/test-runner"]