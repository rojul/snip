FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    fp-compiler \
    && rm -rf /var/lib/apt/lists/*
