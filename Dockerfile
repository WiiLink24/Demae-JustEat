FROM golang:1.23.3-alpine3.19 as builder

# We assume only git is needed for all dependencies.
# openssl is already built-in.
RUN apk add -U --no-cache git

WORKDIR /home/server

# Cache pulled dependencies if not updated.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy necessary parts of the Mail-Go source into builder's source
COPY *.go ./
COPY demae demae
COPY justeat justeat
COPY skip skip

# Build to name "app".
RUN go build -o app .

EXPOSE 4011
CMD ["./app"]
