# Copyright 2020-2021 The OpenEBS Authors. All rights reserved.
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

FROM ubuntu:18.04

LABEL maintainer="OpenEBS"

#Installing necessary ubuntu packages
RUN rm -rf /var/lib/apt/lists/* && \
    apt-get clean && \
    apt-get update --fix-missing || true && \
    apt-get install -y python python-pip netcat iproute2 jq sshpass bc git\
    curl openssh-client

#Installing ansible
RUN pip install ansible==2.7.3

RUN pip install ruamel.yaml.clib==0.1.2

#Installing openshift
RUN pip install openshift==0.11.2

#Installing jmespath
RUN pip install jmespath

RUN touch /mnt/parameters.yml

#Installing Kubectl
ENV KUBE_LATEST_VERSION="v1.20.0"
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl
        
#Adding hosts entries and making ansible folders
RUN mkdir /etc/ansible/ /ansible && \
    echo "[local]" >> /etc/ansible/hosts && \
    echo "127.0.0.1" >> /etc/ansible/hosts

#Copying Necessary Files
COPY ./e2e-tests ./e2e-tests