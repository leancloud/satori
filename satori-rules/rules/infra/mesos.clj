(ns infra.mesos
  (:use riemann.streams
        agent-plugin
        alarm))

(def infra-mesos-rules
  (where (host #"^mesos-slave\d+$")
    (plugin-dir "mesos")

    (where (service "mesos.master.elected")
      (by :host
        (changed :metric {:pairs? true}
          (smap (fn [[ev' ev]] (assoc ev :metric (map #(int (:metric % -1)) [ev' ev])))
            (where (= metric [0 1])
              (! {:note "切换成 Mesos Master 了"
                  :event? true
                  :level 1
                  :groups [:operation]}))
            (where (= metric [1 0])
              (! {:note "不再是 Mesos Master 了"
                  :event? true
                  :level 1
                  :groups [:operation]}))))))

    (where (service "mesos.master.slave_active")
      (by :host
        (set-state-gapped (< 3) (> 4)
          (should-alarm-every 600
            (! {:note "Mesos Slaves 不见了！"
                :level 1
                :expected 5
                :groups [:operation]})))))

    (where (service "mesos.master.uptime_secs")
      (by :host
        (set-state (< 120)
          (changed :state
            (should-alarm-every 60
              (! {:note "Mesos Master 重启了"
                  :level 1
                  :expected 300
                  :groups [:operation]}))))))

    #_(place holder)))
