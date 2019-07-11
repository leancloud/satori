FROM satori:base

WORKDIR /alarm
RUN mkdir /alarm/src
ADD .build/pyproject.toml .build/poetry.lock /alarm/
RUN [ -f /etc/use-mirror ] && (mkdir -p ~/.pip; printf "[global]\nindex-url = https://mirrors.aliyun.com/pypi/simple" > ~/.pip/pip.conf) || true
RUN pip3 install --upgrade poetry
RUN cd /alarm && poetry install

ADD .build/src /alarm/src

EXPOSE 6060

CMD ["/bin/bash", "-c", "cd /alarm && exec poetry run python src/entry.py --config /conf/alarm.yaml"]
