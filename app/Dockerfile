############################
# STEP 1 build executable binary
############################
FROM golang:1.22.1 as builder


WORKDIR /app
COPY go.mod go.sum ./

# curl -v "http://127.0.0.1:7011/health"
RUN apt-get update && apt-get install -y wget ca-certificates gnupg \
&& apt-get -y update \
&& apt-get -y install curl \
&& apt install gettext-base

RUN go mod tidy

COPY . /app

ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o bin /app/cmd

############################
# STEP 2 build a small image
############################
FROM scratch
WORKDIR /app/cmd
# Copy our static executable.
COPY --from=builder /app /app
#COPY --from=builder /app/bin /app/bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ARG SERVICE_PORT=7011
ENV SERVICE__PORT $SERVICE_PORT
EXPOSE $SERVICE_PORT

# Run the app binary.
CMD ["/app/bin"]