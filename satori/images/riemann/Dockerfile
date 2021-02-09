FROM satori:base

WORKDIR /tmp
EXPOSE 5555
ADD app /app
RUN curl http://lc-paas-files.cn-n1.lcfile.com/riemann-0.3.3-satori-standalone.jar -o /app/riemann.jar
CMD ["/app/bootstrap.sh"]
