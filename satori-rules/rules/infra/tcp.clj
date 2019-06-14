(ns infra.tcp
  (:require [riemann.streams :refer :all]
            [agent-plugin :refer :all]
            [alarm :refer :all]
            [lib :refer :all]))

(def infra-tcp-rules
  (->waterfall
    (by :host)
    (where (service "TcpExt.ListenOverflows"))
    (moving-event-window 4)
    (->difference)
    (aggregate max)
    (judge (> 100))
    (changed :state)
    (! {:note "有进程的 Listen 队列爆了"
        :level 5
        :expected 0
        :groups [:operation]})))
