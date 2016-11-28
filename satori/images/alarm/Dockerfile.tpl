# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER bwang@leancloud.rocks

ENV TERM xterm
WORKDIR /alarm
RUN echo "Asia/Shanghai" | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata
RUN adduser ubuntu
RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
RUN apt-get update && apt-get install -y curl python python-dev python-setuptools build-essential git
RUN mkdir /alarm/src
ADD .build/buildout.cfg .build/setup.py .build/docker/use-china-mirror /alarm/
RUN [ -z "USE_MIRROR" ] || /alarm/use-china-mirror
RUN easy_install -U pip zc.buildout setuptools
RUN cd /alarm && buildout

ADD .build/src /alarm/src

EXPOSE 6060

CMD ["/alarm/bin/start", "--config", "/conf/alarm.yaml"]
