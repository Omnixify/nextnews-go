FROM golang:1.26.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /telebot ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 user
USER user
ENV HOME=/home/user
WORKDIR $HOME/app

COPY --from=builder --chown=user:user ./telebot .
EXPOSE 7860

CMD [ "./telebot" ]

