FROM satori:base

WORKDIR /tmp
EXPOSE 5555
ADD app /app
RUN curl http://lc-qjnlgvra.cn-n1.lcfile.com/8bf98b3227aef89cdf44/riemann-0.3.3-satori-standalone.jar -o /app/riemann.jar
CMD ["/app/bootstrap.sh"]
