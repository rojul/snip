FROM nginx:1.13-alpine

WORKDIR /usr/share/nginx/html

COPY nginx.conf /etc/nginx/
COPY --from=snip-web-builder /dist/ ./

HEALTHCHECK CMD [ "wget", "-sq", "http://127.0.0.1/health" ]
