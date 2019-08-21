FROM satori:base

WORKDIR /tmp
RUN curl https://satori.thb.io/grafana_6.1.4_amd64.deb -o grafana-6.1.4.deb && \
    dpkg -i grafana-6.1.4.deb && \
    rm grafana-6.1.4.deb

EXPOSE 3000
WORKDIR /usr/share/grafana
CMD exec /usr/sbin/grafana-server --config=/conf/grafana.ini
