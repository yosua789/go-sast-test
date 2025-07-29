FROM surnet/alpine-wkhtmltopdf:3.21.3-0.12.6-small AS wkhtmltopdf

FROM golang:1.23.5-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./assisttix-api .

##############################################

FROM alpine:3.22 AS final

ENV APP.HOST=0.0.0.0:3000
ENV APP.MODE=prod
ENV APP.DEBUG=false

RUN apk add --no-cache tzdata

RUN apk add --no-cache \
    tzdata \
    runuser \
    libstdc++ \
    libx11 \
    libxrender \
    libxext \
    libssl3 \
    ca-certificates \
    fontconfig \
    freetype \
    ttf-dejavu \
    ttf-droid \
    ttf-freefont \
    ttf-liberation && \
    apk add --no-cache --virtual .build-deps msttcorefonts-installer && \
    update-ms-fonts && \
    fc-cache -f && \
    rm -rf /tmp/* && \
    apk del .build-deps

WORKDIR /app

RUN addgroup assistx && \
    adduser -G assistx -D assistx && \
    chown assistx:assistx /app

USER assistx

COPY --from=wkhtmltopdf /bin/wkhtmltopdf /bin/wkhtmltopdf
COPY --chown=assistx:assistx --from=builder --chmod=744 /app/assisttix-api .
COPY --chown=assistx:assistx --chmod=744 entrypoint.sh entrypoint.sh
COPY --chown=assistx:assistx --from=builder --chmod=744 /app/database/migrations /app/database/migrations

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./assisttix-api"]

