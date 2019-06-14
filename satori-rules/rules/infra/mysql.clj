(ns infra.mysql
  (:require [riemann.streams :refer :all]
            [agent-plugin :refer :all]
            [alarm :refer :all]
            [lib :refer :all]))

(def infra-mysql-rules
  (where (host "mysql-master" "mysql-slave")
    (plugin-dir "mysql")
    (plugin "net.port.listen" 30 {:port 3306})

    (where (and (service "net.port.listen")
                #(= (:port %) 3306))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "MySQL 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))

    #_(place holder)))
