FROM flyio/litefs:0.3.0-beta6 AS litefs

FROM golang AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM debian:bullseye-slim AS tailwindcss
WORKDIR /src

RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl && \
  rm -rf /var/lib/apt/lists/*

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN mv tailwindcss-linux-x64 tailwindcss

COPY tailwind.config.js tailwind.css ./

COPY . ./
RUN ./tailwindcss -i tailwind.css -o app.css --minify

FROM debian:bullseye-slim AS runner
WORKDIR /app

RUN mkdir -p /data /mnt/data

RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates sqlite3 fuse && \
  rm -rf /var/lib/apt/lists/*

ADD litefs.yml /etc/litefs.yml
COPY --from=litefs /usr/local/bin/litefs ./

COPY public ./public/
COPY --from=tailwindcss /src/app.css ./public/styles/
COPY --from=builder /bin/server ./

CMD ["./litefs", "mount"]
