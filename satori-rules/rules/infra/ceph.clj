(ns infra.ceph
  (:use riemann.streams
        agent-plugin
        lib
        alarm))

(def infra-ceph-rules
  (where (host "ceph-host-can-run-ceph-tool")
    (plugin "ceph.status" 10 {})

    (where (service #"^ceph\.mon\.")
      (by :region
        (copy :region :host
          (slot-window :service {:quorum "ceph.mon.quorum"
                                 :total  "ceph.mon.num"}
            (slot-coalesce {:service "ceph.mon.quorum_percent",
                            :metric (/ quorum total)}
              (judge (< 1)
                (runs 2 :state
                  (alarm-every 5 :min
                    (! {:note "Ceph Mon 共识人数不满"
                        :level 3
                        :groups [:operation]}))))
              (judge (< 0.5)
                (runs 2 :state
                  (alarm-every 30 :secs
                    (! {:note "Ceph Mon 无法形成共识"
                        :level 0
                        :groups [:operation]})))))))))

    (where (service #"^ceph\.osd\.")
      (by :region
        (copy :region :host
          (slot-window :service {:total "ceph.osd.num"
                                 :up    "ceph.osd.up"
                                 :in    "ceph.osd.in"}
            (slot-coalesce {:service "ceph.osd.up_percent"
                            :metric (/ up total)}
              (judge (< 1)
                (runs 2 :state
                  (alarm-every 5 :min
                    (! {:note "有不在 UP 状态的 Ceph OSD"
                        :level 3
                        :groups [:operation :devs]})))))
            (slot-coalesce {:service "ceph.osd.in_percent"
                            :metric (/ in total)}
              (judge (< 1)
                (runs 2 :state
                  (alarm-every 5 :min
                    (! {:note "有不在 IN 状态的 Ceph OSD"
                        :level 3
                        :groups [:operation :devs]})))))))))

    (where (and (service "ceph.pg.by_state")
                (#{"degraded" "unknown" "backfilling"} (:pg_state event)))
      (by [:region]
        (copy :region :host
          (where (= (:pg_state event) "degraded")
            (judge (> 0)
              (runs 2 :state
                (alarm-every 30 :min
                  (! {:note "Ceph 中存在降级的 PG"
                      :level 3
                      :groups [:operation :devs]})))))

          (where (= (:pg_state event) "unknown")
            (judge (> 0)
              (runs 2 :state
                (alarm-every 5 :min
                  (! {:note "Ceph 中存在失联的PG（丢失数据！）"
                      :level 1
                      :groups [:operation :devs]})))))

          #_(where (= (:pg_state event) "backfilling")
            (judge (> 0)
              (runs 2 :state
                (alarm-every 3 :hours
                  (! {:note "Ceph 中正在修复降级的 PG"
                      :level 5
                      :groups [:operation :devs]}))))))))))
