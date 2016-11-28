(ns infra.kestrel
  (:use riemann.streams
        agent-plugin
        alarm))

(def ^{:private true} qmeta
  ; "queue-name" [rising falling [:notify :groups]]
  {"some-queue-for-api"  [1000  100  [:operation :api]]
   "some-queue-for-push" [300   50   [:operation :push]]})


(def ^{:private true} qmeta-by-host
  {"some-queue-needs-per-host-alarm" [300   10   [:operation :push]]})

(def ^{:private true} dont-care
  #{"queues-that-does-not-matter"})

(def infra-kestrel-rules
  (where (host #"^kestrel\d+$")
    (plugin-dir "kestrel")
    (plugin-metric "net.port.listen" 30 {:port 22133})

    (where (and (service "net.port.listen")
                (= (:port event) 22133))
      (by :host
        (set-state-gapped (< 1) (> 0)
          (should-alarm-every 120
            (! {:note "kestrel 端口不监听了！"
                :level 0
                :groups [:operation :api]})))))

    (where (and (service "kestrel_queue.items")
                (not (dont-care (:queue event))))

      (let [fgen (fn [i d] #(get-in qmeta [(:queue %) i] d))
            high (fgen 0 200)
            low  (fgen 1 20)
            noti (fgen 2 [:operation])]
        (by [:queue :region]
          (fixed-time-window 60
            (aggregate +
              (copy :region :host
                ; 单调递增报警
                (moving-event-window 10
                  (where (and (>= (count event) 10)
                              (every? pos? (map :metric event)))
                    (aggregate <=
                      (set-state (= true)
                        (should-alarm-every 1800
                          (! {:note #(str "队列 " (:queue %) " 正在堆积！")
                              :level 2
                              :outstanding-tags [:region]
                              :groups noti}))))))

                ; 绝对数量报警
                (set-state-gapped* #(> (:metric %) (high %))
                                   #(< (:metric %) (low %))
                  (runs 3 :state
                    (should-alarm-every 1800
                      (with :description nil
                        (! {:note #(str "队列 " (:queue %) " 爆了！")
                            :level 2
                            :outstanding-tags [:region]
                            :groups noti}))))))))))

      ; 按 host 区分的报警
      (let [fgen (fn [i d] #(get-in qmeta-by-host [(:queue %) i] d))
            has  (fgen 0 nil)
            high (fgen 0 20)
            low  (fgen 1 2)
            noti (fgen 2 [:operation])]
        (where (has (:queue event))
          (by [:host :queue]
            (fixed-time-window 60
              (aggregate +
                (with :descrption nil
                  (set-state-gapped* #(> (:metric %) (high %))
                                     #(< (:metric %) (low %))
                    (runs 3 :state
                      (should-alarm-every 1800
                        (! {:note #(str "队列 " (:queue %) " 爆了！")
                            :level 2
                            :outstanding-tags [:host :queue]
                            :groups noti})))))))))))
    #_(place holder)))
