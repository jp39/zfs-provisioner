FROM docker.io/library/alpine:3.20 as runtime

ENTRYPOINT ["/entrypoint.sh"]

RUN apk add zfs

COPY zfs-provisioner /usr/bin/
COPY docker/entrypoint.sh /
