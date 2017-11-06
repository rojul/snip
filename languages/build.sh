#!/bin/sh
set -e

IMAGE_PREFIX='snip'
TAG='latest'
RUNNER_SNIPPET='
RUN deluser $(getent passwd 1000 | cut -d: -f1); \
    rm -rf /home/*; \
       ( addgroup -g 1000 snip \
      && adduser -DG snip -u 1000 -s /bin/sh -g snip snip) \
    || ( addgroup --gid 1000 snip \
      && adduser --disabled-password --ingroup snip --uid 1000 --shell /bin/sh --gecos snip snip)
WORKDIR /home/snip

COPY --from=snip-runner-builder /runner /usr/local/bin/

ENTRYPOINT []
CMD [ "runner" ]
USER snip
'

info() {
  echo "\033[0;36m$@\033[0m"
}

for dockerfile in */Dockerfile; do
  lang=${dockerfile%/*}
  if [ "$1" -a "$1" != "$lang" ]; then
    continue
  fi
  info Building $lang
  printf "$(cat $dockerfile)$RUNNER_SNIPPET" | docker build -t "$IMAGE_PREFIX/$lang:$TAG" -
done
