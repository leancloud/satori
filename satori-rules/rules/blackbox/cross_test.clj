(ns blackbox.cross-test
  (:use riemann.streams
        agent-plugin
        alarm)
  (:require [clojure.string :as string]))


(defn- ->region [ev]
  (get (string/split (:name ev) #"@") 1))


(defn- app? [ev app]
  (string/starts-with? (:name ev) (str "blackbox-" app "@")))


(def blackbox-cross-test-rules
  (sdo
    (where (host "host-of-cn-n1" "host-of-cn-e1" "host-of-us-w1")
      (plugin-metric "url.check" 30 {:name "blackbox-api@cn-n1", :url "https://api.leancloud.cn/1.1/ping"})
      (plugin-metric "url.check" 30 {:name "blackbox-api@cn-e1", :url "https://e1-api.leancloud.cn/1.1/ping"})
      (plugin-metric "url.check" 30 {:name "blackbox-api@us-w1", :url "https://us-api.leancloud.cn/1.1/ping"})

      #_(plugin-metric "url.check" 30 {:name "blackbox-engine@cn-n1", :url ""})
      #_(plugin-metric "url.check" 30 {:name "blackbox-engine@cn-e1", :url ""})
      #_(plugin-metric "url.check" 30 {:name "blackbox-engine@us-w1", :url ""})

      (plugin-metric "url.check" 30
        {:name "blackbox-push-router@cn-n1",
         :url "https://router-g0-push.leancloud.cn/v1/route?appId=8ezChlx2jBEaHajqsaWfcTnw-gzGzoHsz"})
      (plugin-metric "url.check" 30
        {:name "blackbox-push-router@cn-e1",
         :url "https://router-q0-push.leancloud.cn/v1/route?appId=2ke9qjLSGeamYyU7dT6eqvng-9Nh9j0Va"})
      (plugin-metric "url.check" 30
        {:name "blackbox-push-router@us-w1",
         :url "https://router-a0-push.leancloud.cn/v1/route?appId=mjddqVklO6zYC2rOIr0Nhahp-MdYXbMMI"})

      #_(place holder))


    (where (and (service "url.check.status")
                (string/starts-with? (:name event) "blackbox-"))
      (by [:host :name]
        (adjust [:metric int]
          (moving-event-window 10
            (aggregate* #(->> % (filter (partial not= 200)) count)
              (set-state-gapped (>= 3) (= 0)
                (runs 3 :state
                  (should-alarm-every 120
                    (where (app? event "api")
                      (! {:note #(str "访问 " (->region %) " 的 API 失败率 > 30%")
                          :level 5
                          :expected true
                          :groups [:operation :api]}))
                    (where (app? event "engine")
                      (! {:note #(str "访问 " (->region %) " 的云引擎实例失败率 > 30%")
                          :level 5
                          :expected true
                          :groups [:operation :lean-engine]}))
                    (where (app? event "push-router")
                      (! {:note #(str "访问 " (->region %) " 的 Push Router 失败率 > 30%")
                          :level 5
                          :expected true
                          :groups [:operation :push]}))

                    #_(place holder))))))))

      #_(place holder))))
