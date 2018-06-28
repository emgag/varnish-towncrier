FROM alpine
LABEL maintainer="Matthias Blaser <mb@emgag.com>"
COPY dist/linux_amd64/varnish-towncrier /varnish-towncrier
ENTRYPOINT ["/varnish-towncrier"]