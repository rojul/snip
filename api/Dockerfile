FROM golang:1.9-alpine as builder

RUN apk add --no-cache git \
    && go get github.com/golang/dep/cmd/dep

WORKDIR src/github.com/rojul/snip/api
COPY Gopkg.* ./
RUN dep ensure -vendor-only -v

COPY . ./
RUN go build -o /api -ldflags='-s -w' ./cmd/api

FROM alpine

WORKDIR /root
COPY --from=builder /api .
COPY *.yml ./

EXPOSE 80
HEALTHCHECK CMD [ "wget", "--spider", "-q", "http://127.0.0.1" ]
CMD [ "./api" ]
