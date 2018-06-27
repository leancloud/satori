# FROM USE_MIRRORopenjdk:8
FROM USE_MIRRORjava:8
MAINTAINER feisuzhu@163.com

ENV TERM xterm
ADD .build /app
EXPOSE 6040
EXPOSE 6041
CMD exec /app/master -c /conf/master.yaml
