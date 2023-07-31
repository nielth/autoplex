FROM node:18-alpine AS builder
WORKDIR /app
COPY . .
RUN yarn  
RUN yarn build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html