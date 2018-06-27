# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER feisuzhu@163.com

ENV TERM xterm
WORKDIR /alarm
RUN echo "Asia/Shanghai" | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata
RUN adduser ubuntu
RUN [ -z "USE_MIRROR" ] || (wget http://mirrors.163.com/.help/sources.list.jessie -O /etc/apt/sources.list && rm -rf /etc/apt/sources.list.d/jessie-backports.list)
RUN apt-get update && apt-get install -y curl python git
RUN mkdir /alarm/src
ADD .build/Pipfile .build/Pipfile.lock .build/docker/use-china-mirror .build/docker/get-pip.py /alarm/
RUN [ -z "USE_MIRROR" ] || /alarm/use-china-mirror
RUN python /alarm/get-pip.py
RUN pip install --upgrade pip pipenv
RUN cd /alarm && pipenv install

ADD .build/src /alarm/src

EXPOSE 6060

CMD ["/bin/bash", "-c", "cd /alarm && exec pipenv run python src/entry.py --config /conf/alarm.yaml"]
