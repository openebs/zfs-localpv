# Copyright 2019-2020 The OpenEBS Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM alpine:3.23.4
RUN apk add --no-cache rsyslog openssl xfsprogs xfsprogs-extra ca-certificates
RUN apk add --no-cache btrfs-progs btrfs-progs-extra netcat-openbsd

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
