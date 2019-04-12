FROM openjdk:8-slim
MAINTAINER feisuzhu@163.com
ARG USE_MIRROR

ENV TERM xterm
RUN rm -f /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN adduser ubuntu
RUN [ -z "$USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y curl fcgiwrap supervisor git python redis-server

WORKDIR /tmp
RUN curl https://lc-qjnlgvra.cn-n1.lcfile.com/grafana_6.0.2_amd64.deb -o grafana-6.0.2.deb && \
    dpkg -i grafana-6.0.2.deb && \
    rm grafana-6.0.2.deb

EXPOSE 3000
WORKDIR /usr/share/grafana
CMD exec /usr/sbin/grafana-server --config=/conf/grafana.ini
