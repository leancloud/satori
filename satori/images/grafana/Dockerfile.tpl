# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER bwang@leancloud.rocks

ENV TERM xterm
WORKDIR /tmp
# RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
ADD grafana_3.1.1-1470047149_amd64.deb grafana.deb
RUN dpkg -i grafana.deb
EXPOSE 3000
WORKDIR /usr/share/grafana
CMD exec /usr/sbin/grafana-server --config=/conf/grafana.ini
