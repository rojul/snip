#!/bin/sh
set -e

IMAGE_PREFIX='snip'
TAG='latest'
RUNNER_SNIPPET='
RUN (addgroup -S snip && adduser -SDG snip snip) || adduser --system --group snip \
    && chmod 777 /home/snip
WORKDIR /home/snip

COPY --from=snip-runner /runner /usr/local/bin/

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
