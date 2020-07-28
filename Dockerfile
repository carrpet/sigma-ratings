FROM golang:alpine as builder

WORKDIR $GOPATH/src/github.com/carrpet/sigma-ratings

COPY . .

RUN go build -o /go/bin/sigma-ratings

# Use any base image you need
FROM alpine:latest

RUN apk add --no-cache bash coreutils grep sed
COPY --from=builder /go/bin/sigma-ratings /go/bin/sigma-ratings
ADD appconfig.yml /go/bin/

ENTRYPOINT ["/bin/bash", "-c", "/go/bin/sigma-ratings"]
