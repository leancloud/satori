(ns infra.es
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-es-rules
  (sdo
    (where (host #"^cn-n1-es\d$")
      (plugin "proc.java.heap" 30
        {:name "elasticsearch", :cmdline "org.elasticsearch.bootstrap.Elasticsearch"})


      (where (and (service "proc.java.heap")
                  (= (:name event) "elasticsearch"))
        (by [:host :region]
          (set-state-gapped (> 99.8) (< 95)
            (runs 3 :state
              (should-alarm-every 120
                (! {:note "ElasticSearch OldGen 满了！"
                    :level 1
                    :expected true
                    :outstanding-tags [:host :name]
                    :groups [:operation :api]}))))))

      ; ----------------------------------------

      (plugin "url.check" 30
        {:name "elasticsearch", :url "http://localhost:9200"})

      (where (and (service "url.check.status")
                  (= (:name event) "elasticsearch"))
        (by :host
          (adjust [:metric int]
            (set-state (not= 200)
              (runs 3 :state
                (should-alarm-every 300
                  (! {:note "ElasticSearch 不响应了"
                      :level 1
                      :expected true
                      :groups [:operation :api]}))))))))))
