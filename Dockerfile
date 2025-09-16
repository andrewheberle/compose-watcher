FROM golang:1.24-alpine@sha256:fc2cff6625f3c1c92e6c85938ac5bd09034ad0d4bc2dfb08278020b68540dbb5 AS builder

COPY . /build

RUN cd /build && \
    go build ./cmd/compose-watcher

FROM docker:cli@sha256:0b928cff9f8f13b3475054da4af345db6b21007865f4fa3e0602b4422fea5f99

COPY --from=builder /build/compose-watcher /app/compose-watcher

ENV DOCKER_HOST=unix:///var/run/docker.sock

ENTRYPOINT [ "/app/compose-watcher" ]
CMD [ "watch" ]
