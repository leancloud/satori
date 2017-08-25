(ns leanengine.ping-api
  (:use riemann.streams
        agent-plugin
        alarm))

(def leanengine-ping-api-rules
  (sdo
    (where (host #"^cn-n1-engine\d+$")
      (plugin-metric "url.check" 30
        {:name "api-ping", :url "http://api.leancloud.cn/1.1/ping"}))

    (where (host #"^us-w1-engine\d+$")
      (plugin-metric "url.check" 30
        {:name "api-ping", :url "http://us-api.leancloud.cn/1.1/ping"}))

    (where (host #"^cn-e1-engine\d+")
      (plugin-metric "url.check" 30
        {:name "api-ping", :url "http://e1-api.leancloud.cn/1.1/ping"}))

    (where (and (service "url.check.status")
                (= (:name event) "api-ping"))
      (by :region
        (adjust [:metric int]
          (fixed-time-window 30
            (aggregate* #(apply = 200 %)
              (copy :region :host
                (set-state (= false)
                  (runs 3 :state
                    (should-alarm-every 120
                      (! {:note "云引擎机器上不能访问 API"
                          :level 1
                          :expected true
                          :groups [:operation :lean-engine]}))))))))))

      #_(place holder)))
