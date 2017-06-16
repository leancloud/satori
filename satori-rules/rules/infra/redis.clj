(ns infra.redis
  (:use riemann.streams
        [riemann.index :only [index]]
        agent-plugin
        alarm))

(def infra-redis-rules
  (sdo
    (where (host #"^redis\d+$")
      (plugin-dir "redis")
      (plugin-metric "proc.cpu" 0 {:name "redis", :cmdline "^/usr/bin/redis-server", :interval 5})

      (where (and (service "net.port.listen")
                  (= (:name event) "redis-port"))
        (by [:host :port]
          (set-state (< 1)
            (runs 3 :state
              (should-alarm-every 120
                (! {:note "Redis 端口不监听了"
                    :level 1
                    :expected 3
                    :groups [:operation :api]}))))))

      (where (and (service "proc.cpu")
                  (= (:name event) "redis"))
        (by [:host :name]
          (set-state-gapped (> 85) (< 50)
            (runs 12 :state
              (should-alarm-every 120
                (! {:note "Redis 进程 CPU 占用过高"
                    :level 1
                    :expected 3
                    :groups [:operation :api]}))))))

      (where (service "redis.connected_clients")
        (feed-dog 90 [:port :region])))

    (watchdog
      (where (service "redis.connected_clients")
        (! {:note "Redis 监控数据不上报了！"
            :level 1
            :expected 1
            :outstanding-tags [:port :region]
            :groups [:operation]})))))
