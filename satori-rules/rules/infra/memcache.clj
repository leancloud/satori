(ns infra.memcache
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-memcache-rules
  (where (host #"^cache\d+$")
    (plugin-dir "memcache")
    (plugin-metric "net.port.listen" 30 {:port 11211})

    (where (and (service "net.port.listen")
                #(= (:port %) 11211))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "memcache 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))

    (where (service "memcached.get_hits_ratio")
      (by [:host :port]
        (set-state-gapped (< 80) (> 95)
          (should-alarm-every 120
            (! {:note "memcache 命中率 < 80%"
                :level 2
                :groups [:operation :api]})))))

    (where (service "memcached.curr_connections")
      (by [:host :port]
        (moving-time-window (* 60 5)
          (aggregate maxpdiff
            (set-state-gapped (|>| 0.1) (|<| 0.05)
              (changed :state
                (! {:note #(format "memcache %s 连接数有波动" (:host %))
                    :level 4
                    :groups [:operation :push]})))
            (set-state-gapped (|>| 0.2) (|<| 0.05)
              (changed :state
                (! {:note #(format "memcache %s 连接数抽风了！" (:host %))
                    :level 2
                    :groups [:operation :push]})))))))

    (where (service "memcached.evictions")
      (by [:host :port]
        (moving-time-window (* 60 5)
          (->difference
            (aggregate max
              (set-state-gapped (> 1000) (< 200)
                (changed :state
                  (! {:note "memcache evictions 过高！"
                      :level 2
                      :groups [:operation :api]}))))))))

    #_(place holder)))
