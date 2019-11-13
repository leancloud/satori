FROM satori:base

RUN apt-get install -y libpcre3-dev libssl-dev perl make build-essential zlib1g-dev krb5-multidev libkrb5-dev luarocks && \
    luarocks install lua-resty-auto-ssl && \
    mkdir -p /tmp/build && \
    cd /tmp/build && \
    curl https://openresty.org/download/openresty-1.15.8.2.tar.gz -o openresty.tgz && \
    tar -xzvf openresty.tgz && \
    cd openresty-* && \
    git clone --depth=1 https://github.com/stnoonan/spnego-http-auth-nginx-module && \
    ./configure \
    --with-debug \
    -j$(nproc) \
    --with-pcre-jit \
    `# --with-ipv6` \
    --with-threads \
    --with-file-aio \
    --with-http_v2_module \
    --with-http_realip_module \
    --with-http_addition_module \
    --with-http_gzip_static_module \
    --with-http_auth_request_module \
    --with-http_sub_module \
    --with-http_secure_link_module \
    --with-http_degradation_module \
    --with-http_stub_status_module \
    --with-http_slice_module \
    --with-http_random_index_module \
    --with-stream \
    --with-stream_ssl_module \
    --build=satori-$(date +%Y%m%d) \
    --add-module=spnego-http-auth-nginx-module && \
    make -j $(nproc) && make install

ADD supervisord.conf /etc/supervisord.conf
ADD .build/frontend /frontend
EXPOSE 80
EXPOSE 443

CMD exec supervisord -n -c /etc/supervisord.conf
