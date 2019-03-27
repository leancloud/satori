(ns leanengine.lb
  (:use riemann.streams
        agent-plugin
        alarm))

(def leanengine-lb-rules
  (where (host #"^cn-n1-leanapp-lb\d+$")
    (plugin-dir "haproxy")
    (plugin "net.port.listen" 30 {:port 80  :name "leanapp-lb-80"})
    (plugin "net.port.listen" 30 {:port 443 :name "leanapp-lb-443"})

    (where (and (service "net.port.listen")
                ({"leanapp-lb-80" "leanapp-lb-443"} (:name event)))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "HAProxy 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))

    (where (and (service "haproxy.sratio")
                (= (:proxy-srv event) "FRONTEND"))
      (by :host :proxy
        (set-state-gapped (> 0.9) (< 0.8)
          (should-alarm-every 60
            (! {:note "HAProxy Session 满了！"
                :level 1
                :outstanding-tags [:proxy]
                :groups [:operation :lean-engine]})))))

    #_(place holder)))
