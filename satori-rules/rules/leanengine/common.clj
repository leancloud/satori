(ns leanengine.common
  (:use riemann.streams
        agent-plugin
        alarm))

(def leanengine-common-rules
  (where (host #"docker\d+")
    (plugin-dir "docker")
    (where (service "docker.stuck")
      (by :host
        (set-state (> 0)
          (runs 3 :state
            (should-alarm-every 120
              (! {:note "Docker Daemon 卡死"
                  :level 1
                  :expected 3
                  :groups [:operation :lean-engine]}))))))

    (where (host #"engine\d+$")
      (plugin-metric "tsdb.query" 60
        {:server "opentsdb:4242"
         :name "leanengine.cloud-log-live"
         :expr "len(v[0]['dps'])"
         :m #(str "sum:cloud.log.rpm{host=" (:host %) "}")
         :start "5m-ago"}))

    (where (service "tsdb.query.leanengine.cloud-log-live")
      (by :host
        (set-state (<= 0)
          (should-alarm-every 300
            (! {:note "云引擎日志没有 tsdb 数据点了"
                :level 3
                :expected 1
                :groups [:operation :lean-engine]})))))

    (plugin-metric "proc.num" 30
      {:name "lean-cache-haproxy"
       :cmdline "cache-haproxy.cfg"})

    (where (and (service "proc.num")
                (= "lean-cache-haproxy" (:name event)))
      (by :host
        (set-state (<= 0)
          (runs 3 :state
            (should-alarm-every 300
              (! {:note "LeanCache HAProxy 挂了"
                  :level 1
                  :expected 1
                  :groups [:operation :lean-engine]}))))))

    #_(place holder)))
