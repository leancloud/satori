FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
WORKDIR /tmp
RUN [ -z "USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
ADD grafana_4.3.2_amd64.deb grafana.deb
RUN dpkg -i grafana.deb
EXPOSE 3000
WORKDIR /usr/share/grafana
CMD exec /usr/sbin/grafana-server --config=/conf/grafana.ini
