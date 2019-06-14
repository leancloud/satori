(ns nvgpu.nvgpu
  (:require [clojure.tools.logging :refer [info error]]
            [riemann.streams :refer :all]
            [agent-plugin :refer :all]
            [alarm :refer :all]
            [lib :refer :all]))

(def infra-common-rules
  (sdo
    (where (or (= (:region event) "office")
               (host "host1"
                     "host2"
                     "host3"
                     "host4"))
      (plugin "nvgpu" 0 {:duration 5}))

    (->waterfall
      (where (service "nvgpu.gtemp"))
      (by :host)
      (copy :id :aggregate-desc-key)
      (group-window :id)
      (aggregate max)
      (judge-gapped (> 90) (< 86))
      (alarm-every 2 :min)
      (! {:note "GPU 过热"
          :level 1
          :expected 85
          :outstanding-tags [:host]
          :groups [:operation :devs]}))

    #_(place holder)))
