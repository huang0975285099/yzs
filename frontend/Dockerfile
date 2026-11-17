FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm install --registry=https://registry.npmmirror.com \
    --prefer-offline --no-audit --no-fund --loglevel=http \
    --fetch-timeout=60000 --fetch-retries=2

COPY . .
RUN npm run build

# Final image
FROM nginx:alpine
COPY --from=builder /app/dist/spa /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 18880
CMD ["nginx", "-g", "daemon off;"]
