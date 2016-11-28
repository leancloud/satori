(ns infra.mongodb
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-mongodb-rules
  (where (host #"^mongo\d+")
    (plugin-dir "mongodb")
    (plugin-metric "net.port.listen" 30 {:port 27018})

    (where (service "mongodb.repl.ismaster")
      (by :host
        (changed :metric {:pairs? true}
          (smap (fn [[ev' ev]] (assoc ev :metric (map #(int (:metric % -1)) [ev' ev])))
            (where (= metric [0 1])
              (! {:note "切换成 PRIMARY 了"
                  :event? true
                  :level 1
                  :groups [:operation :api]}))
            (where (= metric [1 0])
              (! {:note "切换成 SECONDARY 了"
                  :event? true
                  :level 1
                  :groups [:operation :api]}))))))

    (where (service "mongodb.connections.available")
      (by :host
        (set-state-gapped (< 10000) (> 50000)
          (should-alarm-every 600
            (! {:note "mongod 可用连接数 < 10000 ！"
                :level 1
                :expected 50000
                :groups [:operation :api]})))))

    (where (service "mongodb.globalLock.currentQueue.total")
      (by :host
        (set-state-gapped (> 250) (< 50)
          (should-alarm-every 600
            (! {:note "MongoDB 队列长度 > 250"
                :level 1
                :expected 50
                :groups [:operation :api]})))))

    #_(place holder)))
