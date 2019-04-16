FROM openjdk:8-slim
MAINTAINER feisuzhu@163.com
ARG USE_MIRROR

ENV TERM xterm
RUN rm -f /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN adduser ubuntu
RUN [ -z "$USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y curl fcgiwrap supervisor git python redis-server

WORKDIR /tmp
EXPOSE 5555
ADD app /app
ADD http://lc-qjnlgvra.cn-n1.lcfile.com/riemann-0.3.1-satori-standalone.jar /app/riemann.jar
CMD ["/app/bootstrap.sh"]
