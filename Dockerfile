FROM alpine:3.5

RUN apk add -U ca-certificates tzdata && rm -Rf /var/cache/apk/*
COPY rt /

ENTRYPOINT ["/rt"]
