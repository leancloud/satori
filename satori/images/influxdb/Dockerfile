FROM openjdk:8-slim
MAINTAINER feisuzhu@163.com
ARG USE_MIRROR

ENV TERM xterm
RUN rm -f /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN adduser ubuntu
RUN [ -z "$USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y curl fcgiwrap supervisor git python redis-server

WORKDIR /tmp
RUN curl https://lc-qjnlgvra.cn-n1.lcfile.com/influxdb_1.7.5_amd64.deb -o influxdb-1.7.5.deb && \
    dpkg -i influxdb-1.7.5.deb && \
    rm influxdb-1.7.5.deb

# Admin server WebUI
EXPOSE 8083

# HTTP API
EXPOSE 8086

# Raft port (for clustering, don't expose publicly!)
#EXPOSE 8090

# Protobuf port (for clustering, don't expose publicly!)
#EXPOSE 8099

VOLUME ["/var/lib/influxdb"]

CMD exec /usr/bin/influxd -config /etc/influxdb/influxdb.conf
