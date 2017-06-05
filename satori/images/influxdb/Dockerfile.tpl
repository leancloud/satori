# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER bwang@leancloud.rocks

ENV TERM xterm
WORKDIR /tmp

ADD influxdb_1.2.4_amd64.deb influxdb.deb
RUN dpkg -i influxdb.deb

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
