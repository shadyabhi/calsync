FROM scratch
COPY calsync /calsync
ENTRYPOINT ["/calsync"]
