(ns infra.cross
  (:require [riemann.streams :refer :all]
            [agent-plugin :refer :all]
            [alarm :refer :all]
            [lib :refer :all]))

(def infra-cross-rules
  (sdo
    ; Ping
    (where (host "office1" "office2")
      (plugin "cross.ping" 15 {:region "office"})
      (plugin "cross.agent" 15 {:region "office"}))

    (where (host "stg1" "stg2")
      (plugin "cross.ping" 15 {:region "stg"})
      (plugin "cross.agent" 15 {:region "stg"}))

    (where (host "prod1" "prod2")
      (plugin "cross.ping" 15 {:region "c1"})
      (plugin "cross.agent" 15 {:region "c1"}))

    (where (service "cross.ping.alive")
      (by :host
        (judge (< 1)
          (runs 3 :state
            (alarm-every 2 :min
              (! {:note "Ping 不通了！"
                  :level 1
                  :expected 1
                  :outstanding-tags [:region]
                  :groups [:operation]}))))))

    (where (service "cross.agent.alive")
      (by :host
        (judge (< 1)
          (runs 3 :state
            (alarm-every 2 :min
              (! {:note "Satori Agent 不响应了！"
                  :level 5
                  :expected 1
                  :outstanding-tags [:region]
                  :groups [:operation]}))))))

    #_(place holder)))
