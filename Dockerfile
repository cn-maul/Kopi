FROM alpine:3.21

WORKDIR /
COPY archiver /archiver

EXPOSE 8082

ENTRYPOINT ["/archiver"]
