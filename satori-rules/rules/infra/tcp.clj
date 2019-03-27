(ns infra.tcp
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-tcp-rules
  (fn [ev]) ; just nop

  #_(by :host
    (where (service "TcpExt.ListenOverflows")
      (moving-event-window 4
        (->difference
          (aggregate max
            (set-state (> 100)
              (changed :state
                (! {:note "有进程的 Listen 队列爆了"
                    :level 5
                    :expected 300
                    :groups [:operation]})))))))))
