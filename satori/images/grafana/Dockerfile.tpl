FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
WORKDIR /tmp
RUN [ -z "USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
ADD http://lc-b5mumjc6.cn-n1.lcfile.com/9e189df8bfe9069d321b.deb grafana-5.2.1.deb
RUN dpkg -i grafana-5.2.1.deb
EXPOSE 3000
WORKDIR /usr/share/grafana
CMD exec /usr/sbin/grafana-server --config=/conf/grafana.ini
