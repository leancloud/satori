FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
RUN rm -f /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN adduser ubuntu
RUN [ -z "USE_MIRROR" ] || sed -E -i 's/(deb|security).debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y curl fcgiwrap supervisor git python redis-server
RUN apt-get install -y libpcre3-dev libssl-dev perl make build-essential zlib1g-dev krb5-multidev libkrb5-dev && \
    mkdir -p /tmp/build && \
    cd /tmp/build && \
    curl https://openresty.org/download/openresty-1.13.6.2.tar.gz -o openresty.tgz && \
    tar -xzvf openresty.tgz && \
    git clone --depth=1 https://github.com/stnoonan/spnego-http-auth-nginx-module && \
    cd openresty-* && \
    ./configure --prefix=/usr --add-module=spnego-http-auth-nginx-module && \
    make -j $(nproc) && make install && \
    cd && rm -rf /tmp/build



ADD supervisord.conf /etc/supervisord.conf
ADD .build/frontend /frontend
EXPOSE 80
EXPOSE 443

CMD exec supervisord -n -c /etc/supervisord.conf
