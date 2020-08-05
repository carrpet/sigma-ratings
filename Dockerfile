FROM golang:alpine as builder

WORKDIR $GOPATH/src/github.com/carrpet/sigma-ratings

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/sigma-ratings 

# Use any base image you need
FROM scratch

COPY --from=builder /go/bin/sigma-ratings /sigma-ratings

EXPOSE 8080

ENTRYPOINT ["/sigma-ratings"]
