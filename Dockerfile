FROM golang:1.25.2
LABEL authors="katagiri"

WORKDIR /avito-test-task

COPY . .

RUN go mod download

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/service/main.go

EXPOSE 8080

CMD ["./app"]