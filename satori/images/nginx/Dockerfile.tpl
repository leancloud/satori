# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER feisuzhu@163.com

ENV TERM xterm
RUN echo "Asia/Shanghai" | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata
RUN adduser ubuntu
RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
RUN apt-get update && apt-get install -y curl nginx fcgiwrap supervisor git

ADD supervisord.conf /etc/supervisord.conf
ADD .build/frontend /frontend
EXPOSE 80
EXPOSE 443

CMD exec supervisord -n -c /etc/supervisord.conf
