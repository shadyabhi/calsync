FROM golang:latest

COPY calsync /calsync
ENTRYPOINT ["/calsync"]
