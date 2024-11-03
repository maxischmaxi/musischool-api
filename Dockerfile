FROM golang:1.23.2-alpine
WORKDIR /app
COPY . .
ENV GIN_MODE=release
RUN go mod download
RUN go build -o main .
CMD ["/app/main"]
