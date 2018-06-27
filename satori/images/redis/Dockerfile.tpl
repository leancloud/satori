# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER feisuzhu@163.com

ENV TERM xterm
RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
RUN apt-get update && apt-get install -y python redis-server
EXPOSE 6379
VOLUME /var/lib/redis
ADD redis.conf /etc/redis/redis.conf
CMD exec /usr/bin/redis-server /etc/redis/redis.conf
