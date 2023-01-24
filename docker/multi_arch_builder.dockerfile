FROM alpine:latest
COPY multi_arch_build.sh /work/multi_arch_build.sh
RUN apk update \
    && apk add --no-cache go bash zip \
    && mkdir -p /build /data

WORKDIR /work
ENTRYPOINT ["/work/multi_arch_build.sh"]