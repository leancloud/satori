(ns infra.nginx
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-nginx-rules
  (where (host #"nginx\d+$")
    (plugin-dir "nginx")
    (plugin-metric "net.port.listen" 30 {:port 80})

    (where (and (service "net.port.listen")
                (= (:port event) 80))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "nginx 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))

    (where (service "nginx.upstream.healthy.ratio")
      (by [:host :upstream]
        (set-state-gapped (< 0.2) (> 0.8)
          (should-alarm-every 120
            (! {:note #(str (:upstream %) " upstream 死干净了！")
                :level 0
                :groups [:operation :api]})))))

    #_(place holder)))
