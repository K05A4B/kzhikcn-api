FROM golang:1.25-bookworm AS build

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .
COPY . .

RUN apt-get update && apt-get install -y gcc
RUN go mod download

RUN go build -o ./kzhikcn

# base：公共运行时。安装 tini 以保证信号转发。
FROM debian:bookworm-slim AS base

WORKDIR /app

COPY --from=build /app/kzhikcn .

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tini \
 && rm -rf /var/lib/apt/lists/*

RUN mkdir -p ./sys ./data \
 && /app/kzhikcn -c /app/sys/config.yml gen-config -d \
 && ln -s /app/sys/config.yml /app/config.yml

EXPOSE 5083

ENTRYPOINT ["/usr/bin/tini", "--"]
# 通过 exec 让 Go 进程直接接收信号；ADDRESS 可覆盖监听地址
CMD ["/bin/sh", "-c", "exec ./kzhikcn serve -a ${ADDRESS:-0.0.0.0:5083}"]

# production：精简镜像，仅包含运行时依赖
FROM base AS production

# extends：在精简镜像基础上内置 curl / jq / python 等常用工具，
# 便于 events 的 command 钩子执行脚本与网络请求。
FROM base AS extends

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    jq \
    bash \
    python3 \
    python3-pip \
    python3-venv \
    python-is-python3 \
    coreutils \
    findutils \
    grep \
    sed \
    gawk \
    tar \
    gzip \
    unzip \
    git \
    openssl \
    netcat-openbsd \
    dnsutils \
    tzdata \
    imagemagick \
 && rm -rf /var/lib/apt/lists/*
