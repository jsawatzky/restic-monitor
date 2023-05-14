FROM golang:1.20-alpine AS build

WORKDIR /go/src
COPY . .

RUN go get -d -v ./...
RUN go build -a -o restic-monitor .

FROM restic/restic:0.15.1 AS runtime

RUN mkdir /.restic-cache
VOLUME /.restic-cache

ENV RESTIC_CACHE_DIR=/.restic-cache

COPY --from=build /go/src/restic-monitor /usr/local/bin/restic-monitor

ENTRYPOINT [ "/usr/local/bin/restic-monitor" ]