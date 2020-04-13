FROM golang:1.14-alpine as builder
RUN mkdir -p $GOPATH/src/app
ADD . $GOPATH/src/app
WORKDIR $GOPATH/src/app
RUN apk add --no-cache ca-certificates
ENV CGO_ENABLED 0
ENV GOOS linux
# Build the binary
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM scratch
# ca-certificates are required to access external https services
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# copy the application binary
COPY --from=builder /go/src/app/main /app/
WORKDIR /app
ENTRYPOINT ["./main"]