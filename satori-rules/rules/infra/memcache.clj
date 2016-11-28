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
    #_(place holder)))
