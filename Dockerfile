FROM docker.io/library/alpine:3.11 as runtime

ENTRYPOINT ["kubernetes-zfs-provisioner"]

RUN \
    apk add --no-cache curl bash openssh && \
    adduser -S zfs -G root

COPY packaging/zfs.sh /usr/bin/zfs
COPY kubernetes-zfs-provisioner /usr/bin/

USER zfs:root
