FROM golang:1.23-alpine

RUN apk add --no-cache git postgresql-client

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ozon-comments-graphql

CMD while ! pg_isready -h db -U postgres; do sleep 2; done && ./ozon-comments-graphql