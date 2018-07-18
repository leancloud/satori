FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
WORKDIR /tmp
ADD http://lc-b5mumjc6.cn-n1.lcfile.com/034fa6024b475236060f.deb influxdb-1.6.0.deb
RUN dpkg -i influxdb-1.6.0.deb

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
