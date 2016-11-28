# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER bwang@leancloud.rocks

ENV TERM xterm
WORKDIR /tmp
RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
RUN apt-get update && apt-get -y install git supervisor
EXPOSE 5555
ADD app /app
ADD .build/riemann-reloader /app/riemann-reloader
CMD ["/usr/bin/supervisord","-n","-c","/app/supervisord.conf"]
