FROM alpine
LABEL org.opencontainers.image.source=https://github.com/emgag/varnish-towncrier
COPY varnish-towncrier /varnish-towncrier
ENTRYPOINT ["/varnish-towncrier"]
