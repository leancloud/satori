(ns infra.apilog-es
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-apilog-es-rules
  (where (host #"^host-of-elasticsearch\d$")
    (plugin-metric "url.check" 30
      {:name "cn-apilog-es-ping", :url "http://localhost:9200"})

    (where (and (service "url.check.status")
                (= (:name event) "cn-apilog-es-ping"))
      (by :host
        (adjust [:metric int]
          (set-state (not= 200)
            (runs 3 :state
              (should-alarm-every 300
                (! {:note "访问日志 ElasticSearch 不响应了"
                    :level 3
                    :expected true
                    :groups [:operation :api]})))))))

      #_(place holder)))
