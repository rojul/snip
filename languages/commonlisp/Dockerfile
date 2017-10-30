FROM debian:stable-slim

ENV CLOZURE_VERSION=1.11 \
    PATH=$PATH:/usr/local/src/ccl:/usr/local/src/ccl/scripts

RUN buildDeps='wget' \
    && apt-get update && apt-get install -y --no-install-recommends $buildDeps \
    && rm -rf /var/lib/apt/lists/* \
    && wget --no-verbose -O ccl.tar.gz ftp://ftp.clozure.com/pub/release/$CLOZURE_VERSION/ccl-$CLOZURE_VERSION-linuxx86.tar.gz \
    && mkdir -p /usr/local/src \
    && tar -xf ccl.tar.gz -C /usr/local/src \
    && rm -rf /usr/local/src/ccl/.svn ccl.tar.gz \
    && apt-get purge -y --auto-remove $buildDeps
