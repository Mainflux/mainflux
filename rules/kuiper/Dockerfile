FROM golang:1.15.1-alpine AS builder

RUN apk add gcc make git libc-dev binutils-gold pkgconfig zeromq-dev
RUN git clone https://github.com/lf-edge/ekuiper /go/kuiper

WORKDIR /go/kuiper

# run in /go/kuiper in order to use mainflux versions of packages
RUN go get github.com/mainflux/mainflux

# mainflux plugins
RUN git clone https://github.com/mainflux/mainflux
RUN cp mainflux/rules/kuiper/source.go extensions/sources/mainflux.go \
    && go build -trimpath -modfile extensions.mod --buildmode=plugin -o extensions/sources/mainflux.so extensions/sources/mainflux.go
RUN cp mainflux/rules/kuiper/sink.go extensions/sinks/mainflux.go \
    && go build -trimpath -modfile extensions.mod --buildmode=plugin -o extensions/sinks/mainflux.so extensions/sinks/mainflux.go

# kuiper
RUN go build -o kuiper_conf_util deploy/docker/conf_util.go \
    && if [ "$(uname -m)" = "x86_64" ]; then apk add upx && upx ./kuiper_conf_util; fi
RUN make build_without_edgex

FROM alpine:3.12

# Set environment vars
ENV MAINTAINER="emqx.io" \
    KUIPER_HOME="/kuiper" \
    KUIPER__BASIC__CONSOLELOG=true

# These vars are not persisted in the final image layer
ARG KUIPER_USER="kuiper"
ARG KUIPER_USER_ID="1001"

# (root) Add packages and "kuiper" user
RUN apk add sed libzmq

COPY --from=builder /go/kuiper/deploy/docker/docker-entrypoint.sh /usr/bin/docker-entrypoint.sh
COPY --from=builder /go/kuiper/kuiper_conf_util /usr/bin/kuiper_conf_util
COPY --from=builder /go/kuiper/_build/kuiper-* /kuiper/

# mainflux plugins
COPY --from=builder /go/kuiper/extensions/sources/mainflux.so /kuiper/plugins/sources/mainflux.so
COPY --from=builder /go/kuiper/extensions/sinks/mainflux.so /kuiper/plugins/sinks/mainflux.so

# Set appropriate ownership to allow binary full access to KUIPER_HOME dir
RUN adduser -DH -s /sbin/nologin -u ${KUIPER_USER_ID} ${KUIPER_USER} && \
    chown -Rh ${KUIPER_USER}:${KUIPER_USER} ${KUIPER_HOME} && \
    mkdir -p /usr/local/taos && \
    chown -Rh ${KUIPER_USER}:${KUIPER_USER} /usr/local/taos

# Run the kuiper process under the kuiper user
USER ${KUIPER_USER}

VOLUME ["${KUIPER_HOME}/etc", "${KUIPER_HOME}/data", "${KUIPER_HOME}/plugins", "${KUIPER_HOME}/log"]
EXPOSE 9081 20498

WORKDIR ${KUIPER_HOME}

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

CMD ["./bin/kuiperd"]
