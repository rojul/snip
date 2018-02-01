FROM abiosoft/caddy:builder as builder

ARG version="0.10.10"
ARG plugins=""

RUN VERSION=${version} PLUGINS=${plugins} /bin/sh /usr/bin/builder.sh \
    && /install/caddy -version \
    && /install/caddy -plugins

FROM alpine:3.7

RUN apk add --no-cache ca-certificates

COPY --from=builder /install/caddy /usr/bin/caddy

VOLUME /root/.caddy
WORKDIR /srv

ENTRYPOINT [ "caddy", "--conf", "/etc/Caddyfile", "--log", "stdout" ]

COPY Caddyfile /etc/
COPY --from=snip-web-builder /dist/ /srv/

EXPOSE 80 443
