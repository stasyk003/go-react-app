FROM golang:1.20.2 as builder

WORKDIR /app

COPY . .

RUN go mod download && go get -u ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# multi-stage build
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app .

ENV MONGODB_URI=mongodb://host.docker.internal:27017

EXPOSE 8000

CMD ["./main"]


