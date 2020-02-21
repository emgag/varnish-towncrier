FROM alpine
LABEL maintainer="Matthias Blaser <mb@emgag.com>"
COPY varnish-towncrier /varnish-towncrier
ENTRYPOINT ["/varnish-towncrier"]
