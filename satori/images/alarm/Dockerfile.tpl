FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
RUN rm -f /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN adduser ubuntu
RUN [ -z "USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y curl nginx fcgiwrap supervisor git python redis-server

WORKDIR /alarm
RUN mkdir /alarm/src
ADD .build/Pipfile .build/Pipfile.lock .build/docker/use-china-mirror .build/docker/get-pip.py /alarm/
RUN [ -z "USE_MIRROR" ] || /alarm/use-china-mirror
RUN python /alarm/get-pip.py
RUN pip install --upgrade pip!=18.1 pipenv
RUN cd /alarm && pipenv install

ADD .build/src /alarm/src

EXPOSE 6060

CMD ["/bin/bash", "-c", "cd /alarm && exec pipenv run python src/entry.py --config /conf/alarm.yaml"]
