# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER feisuzhu@163.com

ENV TERM xterm
ADD .build /app
EXPOSE 8433
CMD exec /app/transfer -c /conf/transfer.yaml
