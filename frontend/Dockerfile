# syntax=docker/dockerfile:1

ARG NODE_VERSION=20
ARG DEBIAN_VERSION=bookworm

FROM node:${NODE_VERSION}-${DEBIAN_VERSION} AS builder
ARG NPM_REGISTRY=https://registry.npmmirror.com
ARG APT_MIRROR=mirrors.tuna.tsinghua.edu.cn
WORKDIR /app
RUN sed -i "s/deb.debian.org/${APT_MIRROR}/g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i "s/deb.debian.org/${APT_MIRROR}/g" /etc/apt/sources.list || true
COPY package.json package-lock.json ./
RUN npm config set registry "${NPM_REGISTRY}" && npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=3s --retries=10 CMD wget -qO- http://127.0.0.1/ || exit 1
