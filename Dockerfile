FROM golang:alpine AS builder

RUN mkdir /tmp
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" .

FROM scratch

# Create the /tmp directory in the final image.
COPY --from=builder /tmp /tmp

EXPOSE 22
ENTRYPOINT ["/fakessh"]
