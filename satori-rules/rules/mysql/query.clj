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
         :sql "SELECT count(*) FROM user WHERE bad = 1"})

      (plugin-metric "mysql.query" 60
        {:name "total-user-count"
         :host "localhost"
         :port 3306
         :user "user_for_monitor"
         :password "Secret!IMeanIt!"
         :database "awesome_app"
         :sql "SELECT count(*) FROM user"})))

    (where (service "mysql.query.bad-user-count")
      (set-state (> 50)
        (runs 2 :state
          (should-alarm-every 300
            (! {:note "坏用户太多了！"
                :level 3
                :groups [:operation]}))))

    (where (service #"^mysql\.query\..*-user-count$")
      (slot-window :service {:total "mysql.query.total-user-count",
                             :bad   "mysql.query.bad-user-count"}
        (slot-coalesce {:service "app.sms.bad-user-percent",
                        :metric (if (> bad 5) (/ bad total) -1)}
          (set-state (> 0.5)
            (runs 2 :state
              (should-alarm-every 300
                (! {:note "坏用户占比过高！"
                    :level 3
                    :groups [:operation]})))))))))
