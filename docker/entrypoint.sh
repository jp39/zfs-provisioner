#!/bin/sh

rpcbind
rpc.statd --no-notify --port 32765 --outgoing-port 32766
rpc.mountd --port 32767
rpc.idmapd
rpc.nfsd --tcp --udp --port 2049 8

exec /usr/bin/kubernetes-zfs-provisioner
