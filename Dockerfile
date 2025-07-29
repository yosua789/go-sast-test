FROM golang:1.23.5-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./assisttix-api .

##############################################

FROM alpine:3.20 AS final

ENV APP.HOST=0.0.0.0:3000
ENV APP.MODE=prod
ENV APP.DEBUG=false

RUN apk add --no-cache tzdata

WORKDIR /app

RUN addgroup assistx && \
    adduser -G assistx -D assistx && \
    chown assistx:assistx /app

USER assistx

COPY --chown=assistx:assistx --from=builder --chmod=744 /app/assisttix-api .
COPY --chown=assistx:assistx --chmod=744 entrypoint.sh entrypoint.sh
COPY --chown=assistx:assistx --from=builder --chmod=744 /app/database/migrations /app/database/migrations

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./assisttix-api"]

