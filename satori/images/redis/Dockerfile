FROM satori:base

EXPOSE 6379
ADD redis.conf /etc/redis/redis.conf
CMD exec /usr/bin/redis-server /etc/redis/redis.conf
