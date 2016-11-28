(ns infra.es
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-es-rules
  (where (host #"^es\d$")
    (plugin-metric "proc.java.heap" 30
      {:name "cn-api-es", :cmdline "org.elasticsearch.bootstrap.Elasticsearch"})

    (where (and (service "proc.java.heap")
                (= (:name event) "cn-api-es"))
      (by [:host :name]
        (set-state-gapped (> 99.8) (< 95)
          (runs 3 :state
            (should-alarm-every 120
              (! {:note "应用内搜索 ElasticSearch OldGen 满了！"
                  :level 1
                  :expected true
                  :outstanding-tags [:host :name]
                  :groups [:operation :api]}))))))

      #_(place holder)))
