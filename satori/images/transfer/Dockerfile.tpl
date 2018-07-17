FROM USE_MIRRORopenjdk:8-slim
MAINTAINER feisuzhu@163.com

ENV TERM xterm
ADD .build /app
EXPOSE 8433
CMD exec /app/transfer -c /conf/transfer.yaml
