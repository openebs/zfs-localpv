FROM alpine:3.23.5
RUN apk add --no-cache util-linux
RUN apk add --no-cache btrfs-progs btrfs-progs-extra xfsprogs xfsprogs-extra e2fsprogs e2fsprogs-extra
RUN apk add --no-cache ca-certificates rsyslog openssl netcat-openbsd

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL

COPY zfs-driver /usr/local/bin/

LABEL org.label-schema.name="zfs-driver"
LABEL org.label-schema.description="OpenEBS ZFS LocalPV Driver"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["/usr/local/bin/zfs-driver"]
EXPOSE 7676 7777
