(ns mysql.query
  (:use riemann.streams
        agent-plugin
        alarm)
  (:require [clojure.string :as string]))

(def mysql-query-rules
  (sdo
    (where (host "db-sms02")
      ; 更多用法看源码： plugins/_metric/mysql.query
      ; 要注意安全问题
      (plugin-metric "mysql.query" 60
        {:name "bad-user-count"
         :host "localhost"
         :port 3306
         :user "user_for_monitor"
         :password "Secret!IMeanIt!"
         :database "awesome_app"
         :sql "SELECT count(*) FROM user WHERE bad = 1"}))

    (where (service "mysql.query.bad-user-count")
      (set-state (> 50)
        (runs 2 :state
          (should-alarm-every 300
            (! {:note "坏用户太多了！"
                :level 3
                :groups [:operation]})))))))
