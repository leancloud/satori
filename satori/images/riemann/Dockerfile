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
RUN curl http://lc-qjnlgvra.cn-n1.lcfile.com/8bf98b3227aef89cdf44/riemann-0.3.3-satori-standalone.jar -o /app/riemann.jar
CMD ["/app/bootstrap.sh"]
