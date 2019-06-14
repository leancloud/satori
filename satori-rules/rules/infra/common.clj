(ns infra.common
  (:require [riemann.streams :refer :all]
            [agent-plugin :refer :all]
            [alarm :refer :all]
            [lib :refer :all]
            [hostgroup :refer :all]
            [clojure.tools.logging :refer [info error]]))

(def infra-common-rules
  (sdo
    (where (service "agent.alive")
      ; HACK: satori-agent in containers will not
      ;       report builtin metrics,
      ;       by doing this these config will not assign to agents in containers.
      (plugin-dir "infra")
      (plugin "proc.num" 30 {:cmdline "^/usr/sbin/ntpd " :name "proc-ntpd"}))

    (where (and (service "proc.num")
                (= (:name event) "proc-ntpd"))
      (by :host
        (judge (< 1)
          (alarm-every 2 :min
            (! {:note "NTP 进程不在了"
                :level 5
                :expected 1.0
                :groups [:operation]})))))

    (where (service "mem.swaponfile")
      (by :host
        (judge (> 0)
          (alarm-every 550 :secs
            (! {:note "有基于文件的 swap"
                :level 5
                :expected 0
                :groups [:operation]})))))

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
        (judge-gapped (> 1.5) (< 1.2)
          (runs 3 :state
            (changed :state
              (! {:note "Load 150%"
                  :level 5
                  :expected 1.0
                  :groups [:operation]}))))
        (judge-gapped (> 2.0) (< 1.8)
          (runs 3 :state
            (changed :state
              (! {:note "Load 200%"
                  :level 3
                  :expected 1.5
                  :groups [:operation]}))))
        (judge-gapped (> 3.0) (< 2.5)
          (runs 3 :state
            (changed :state
              (! {:note "Load 300%"
                  :level 1
                  :expected 2.0
                  :groups [:operation]}))))))

    ; 磁盘
    (where (service "df.bytes.used.percent")
      (by [:host :mount]
        (judge-gapped (> 80) (< 75)
          (alarm-every 5 :hours
            (! {:note "磁盘80%"
                :level 5
                :expected 85
                :groups [:operation]})))
        (judge-gapped (> 90) (< 85)
          (alarm-every 10 :min
            (! {:note "磁盘90%（要满了！）"
                :level 1
                :expected 95
                :groups [:operation]})))))

    (where (service "df.inodes.used.percent")
      (by [:host :mount]
        (judge-gapped (> 80) (< 75)
          (alarm-every 2 :hours
            (! {:note "inode 80%"
                :level 3
                :expected 85
                :groups [:operation]})))))

    (where (service "megaraid.offline")
      (by :host
        (judge (> 0)
          (alarm-every 10 :min
            (! {:note "RAID 里有盘坏掉了！"
                :level 1
                :expected 0
                :groups [:operation]})))))

    ; CPU
    (where (service "cpu.idle")
      (by :host
        (judge-gapped (< 30) (> 50)
          (runs 5 :state
            (alarm-every 15 :min
              (! {:note "CPU 占用过高！"
                  :level 3
                  :expected 30
                  :groups [:operation]}))))))

    (where (service "cpu.steal")
      (by :host
        (judge-gapped (> 30) (< 5)
          (runs 2 :state
            (alarm-every 15 :min
              (! {:note "CPU 被偷了！"
                  :level 5
                  :expected 0
                  :groups [:operation]}))))))

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

    (where (and (service "agent.ping")
                (not (host "forum")))
      (by :host
        (set-state (< 1)
          (runs 3 :state
            (should-alarm-every 120
              (! {:note "Satori Agent 不响应了！"
                  :level 5
                  :expected 1
                  :outstanding-tags [:region]
                  :groups host->group}))))))

    ; Zombies
    (where (service "proc.zombies")
      (by :host
        (judge (> 15)
          (runs 3 :state
            (alarm-every 2 :hours
              (! {:note "有僵尸进程"
                  :level 6
                  :expected 15
                  :groups [:operation]}))))))

    (where (service "proc.uninterruptables")
      (by :host
        (judge (> 30)
          (runs 5 :state
            (alarm-every 2 :hours
              (! {:note "有好多 D 状态的进程"
                  :level 2
                  :expected 10
                  :groups [:operation]}))))))

    (where (service "net.netfilter.conntrack.used_ratio")
      (by :host
        (judge-gapped (> 0.85) (< 0.75)
          (runs 5 :state
            (alarm-every 5 :min
              (! {:note "conntrack 要满了"
                  :level 2
                  :expected 10
                  :groups [:operation]}))))))

    ; Kernel BUG
    #_(where (service "kernel.dmesg.bug")
      (by :host
        (judge (> 1)
          (changed :state
            (! {:note "Kernel BUG"
                :level 1
                :expected 3
                :groups [:operation]})))))

    ; IO Error
    #_(where (service "kernel.dmesg.io_error")
      (by :host
        (judge (> 1)
          (changed :state
            (! {:note "有磁盘错误"
                :level 1
                :expected 3
                :groups [:operation]})))))

    #_(place holder)))
