FROM satori:base

ADD .build /app
EXPOSE 6040
EXPOSE 6041
CMD exec /app/master -c /conf/master.yaml
