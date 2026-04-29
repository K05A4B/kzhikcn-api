FROM golang:1.24-bookworm AS build

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .
COPY . .

RUN apt-get update && apt-get install -y gcc
RUN go mod download

RUN go build -o ./kzhikcn

FROM debian:bookworm-slim AS production

WORKDIR /app

COPY --from=build /app/kzhikcn .

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*


EXPOSE 5083
CMD ["./kzhikcn", "serve", "-a", "${ADDRESS}"]