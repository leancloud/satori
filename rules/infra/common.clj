(ns infra.common
  (:use riemann.streams
        hostgroup
        agent-plugin
        alarm))

; host->group 这个函数在 hostgroup.clj 里，模仿 Open-Falcon 的 HostGroup 机制
(def infra-common-rules
  (sdo
    (plugin-dir "infra")

    ; agent 上报
    #_(where (service "agent.alive")
      (feed-dog 90))

    #_(watchdog
      (where (service "agent.alive")
        (! {:note "Agent.Alive 不上报了！"
            :level 2
            :expected 1
            :outstanding-tags [:region]
            :groups host->group})))

    ; 机器 Load
    (where (service "load.15min.normalized")
      (by :host
        (set-state-gapped (> 1.5) (< 1.2)
          (runs 3 :state
            (changed :state
              (! {:note "Load 150%"
                  :level 5
                  :expected 1.0
                  :groups host->group}))))
        (set-state-gapped (> 2.0) (< 1.8)
          (runs 3 :state
            (changed :state
              (! {:note "Load 200%"
                  :level 3
                  :expected 1.5
                  :groups host->group}))))
        (set-state-gapped (> 3.0) (< 2.5)
          (runs 3 :state
            (changed :state
              (! {:note "Load 300%"
                  :level 1
                  :expected 2.0
                  :groups host->group}))))))

    ; 磁盘
    (where (service "df.bytes.used.percent")
      (by [:host :mount]
        (set-state-gapped (> 90) (< 85)
          (should-alarm-every 7200
            (! {:note "磁盘90%"
                :level 3
                :expected 85
                :groups host->group})))
        (set-state-gapped (> 98) (< 95)
          (should-alarm-every 600
            (! {:note "磁盘98%（要满了！）"
                :level 1
                :expected 95
                :groups host->group})))))

    (where (service "df.inodes.used.percent")
      (by [:host :mount]
        (set-state-gapped (> 90) (< 85)
          (should-alarm-every 7200
            (! {:note "inode 90%"
                :level 3
                :expected 85
                :groups host->group})))))

    (where (service "megaraid.offline")
      (by :host
        (set-state (> 0)
          (should-alarm-every 600
            (! {:note "RAID 里有盘坏掉了！"
                :level 1
                :expected 0
                :groups host->group})))))

    ; CPU
    (where (service "cpu.idle")
      (by :host
        (set-state-gapped (< 15) (> 20)
          (runs 5 :state
            (should-alarm-every 900
              (! {:note "CPU 占用过高！"
                  :level 3
                  :expected 30
                  :groups host->group}))))))

    ; Ping
    ; 这里需要看 Ping 插件。Ping 插件是从 PuppetDB 中取机器列表的，请根据自己的需求修改。
    (where (host "hosts" "which" "perform" "ping")
      (plugin-dir "ping"))

    (where (service "ping.alive")
      (by :host
        (set-state (< 1)
          (runs 3 :state
            (should-alarm-every 120
              (! {:note "Ping 不通了！"
                  :level 1
                  :expected 1
                  :outstanding-tags [:region]
                  :groups host->group}))))))

    ; Zombies
    (where (service "proc.zombies")
      (by :host
        (set-state (> 15)
          (runs 3 :state
            (should-alarm-every 7200
              (! {:note "有僵尸进程"
                  :level 6
                  :expected 15
                  :groups host->group}))))))

    (where (service "proc.uninterruptables")
      (by :host
        (set-state (> 30)
          (runs 5 :state
            (should-alarm-every 7200
              (! {:note "有好多 D 状态的进程"
                  :level 2
                  :expected 10
                  :groups host->group}))))))

    (where (service "net.netfilter.conntrack.used_ratio")
      (by :host
        (set-state-gapped (> 0.85) (< 0.75)
          (runs 5 :state
            (should-alarm-every 300
              (! {:note "conntrack 要满了"
                  :level 2
                  :expected 10
                  :groups host->group}))))))

    ; Kernel BUG
    #_(where (service "kernel.dmesg.bug")
      (by :host
        (set-state (> 1)
          (changed :state
            (! {:note "Kernel BUG"
                :level 1
                :expected 3
                :groups host->group})))))

    ; IO Error
    #_(where (service "kernel.dmesg.io_error")
      (by :host
        (set-state (> 1)
          (changed :state
            (! {:note "有磁盘错误"
                :level 1
                :expected 3
                :groups host->group})))))

    #_(place holder)))
