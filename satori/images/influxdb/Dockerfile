FROM satori:base

WORKDIR /tmp

RUN curl https://satori.thb.io/influxdb_1.6.3_amd64.deb -o influxdb-1.6.3.deb && \
    dpkg -i influxdb-1.6.3.deb && \
    rm influxdb-1.6.3.deb

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
