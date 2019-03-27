(ns infra.redis
  (:use riemann.streams
        [riemann.index :only [index]]
        agent-plugin
        alarm))

(def infra-redis-rules
  (sdo
    (where (host #"^redis\d+$")
      (plugin-dir "redis")
      (plugin "proc.cpu" 0 {:name "redis", :cmdline "^/usr/bin/redis-server", :interval 5})

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

      (where (service #"^redis\.")
        (by [:host :port :region]
          (slot-window :service {:curmem "redis.used_memory"
                                 :maxmem "redis.maxmemory"}
            (slot-coalesce {:service "redis.memory_used_ratio"
                            :metric (if (pos? maxmem) (/ curmem 0.01 maxmem) 0)
                            :port (:port ev:curmem)
                            :region (:region ev:curmem)
                            :curmem curmem
                            :maxmem maxmem}
              (set-state (> 95)
                (runs 2 :state
                  (should-alarm-every 60
                    (! {:note "Redis 内存占用快满了！"
                        :level 3
                        :expected 75
                        :outstanding-tags [:host :port :region]
                        :groups [:operation :api :push]})))))))))))

(defn redis-sentinel
  [hostname sname url]
  (where (and (host hostname)
              (service "agent.alive"))
    (plugin "redis.sentinel" 30 {:name sname, :url url})))


(def infra-redis-sentinel-rules
  (sdo
    (redis-sentinel #"^sentinel[1-5]$" "redis-api"  "redis://localhost:26380")
    (redis-sentinel #"^sentinel[1-5]$" "redis-feed" "redis://localhost:26381")

    (where (service "redis_sentinel.status")
      (by [:host :name :master]
        (set-state (> 0)
          (should-alarm-every 60
            (! {:note "Redis Sentinel 状态不正确！"
                :level 5
                :expected 0
                :outstanding-tags [:host :name :master]
                :groups [:operation]})))))))
