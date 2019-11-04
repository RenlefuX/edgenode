# Copyright 2019 Intel Corporation. All rights reserved
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

FROM centos:7.6.1810

RUN yum install -y pciutils which unbound openssl iproute \
    && yum clean all
RUN rpm -ivh https://github.com/alauda/ovs/releases/download/v2.11.4-1/openvswitch-2.11.4-1.el7.x86_64.rpm

WORKDIR /root/

COPY ./interfaceservice .
COPY ./entrypoint_interfaceservice.sh .

CMD ["./entrypoint_interfaceservice.sh"]