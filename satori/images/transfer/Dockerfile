FROM satori:base

ADD .build /app
EXPOSE 8433
CMD exec /app/transfer -c /conf/transfer.yaml
