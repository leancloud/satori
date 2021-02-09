(ns alarm
  (:require [clojure.data.json :as json]
            [clojure.set :as cset]
            [clojure.string :as string]
            [riemann.common :refer [event]]
            [riemann.config :refer :all]
            [riemann.streams :refer :all]
            [riemann.test :refer [io]]
            [riemann.time :refer [unix-time]]
            [taoensso.carmine :as car]
            [clojure.tools.logging :refer [info]])

  (:import [riemann.codec Event]
           [org.quartz CronExpression]))

(defn- evid [ev outstanding-tags m]
  (let [f #(let [v (% m)] (if (fn? v) (v ev) v))]
    (->> outstanding-tags
        (map #(str %1 "=" (%1 ev)))
        (string/join ",")
        (str (:host ev) "/"
             (:service ev) "/"
             (f :level) "/"
             (f :note) "/")
        (hash)
        (format "%x"))))

(defn- all-tags [ev]
  (-> (cset/difference (set (keys ev)) (set (map keyword (Event/getBasis)))) (sort) (vec)))

(def ^{:private true} alarm-url (atom "redis://localhost:6379"))

(defn set-alarm-redis! [url]
  (reset! alarm-url url))

(defn !
  "
  创建一个发送报警的流，流经这个流的事件都会被发到 alarm 产生报警。
  接受一个 map 做参数，map 中需要可以指定如下的参数

  (! {:note string ; 报警标题，标题对于一个特定的报警是不能变的（不要把报警的数据编码在这里面）
      :level 1  ;报警级别, 0最高，6最小。报警级别影响报警方式。
      :event? false  ; 可选，是不是事件（而不是状态）。默认 false。如果是事件的话，只会发报警，不会记录状态（alarm插件里看不到）。
      :outstanding-tags [:region :mount]  ; 可选，相关的tag，写在这里的 tag 会用于区分不同的事件，以及显示在报警内容中, 不填的话默认是所有的tag
      :groups [:operation]  ; groups 是在规则仓库的 alarm 配置里管理的)
      :meta {:foo 1 :bar 2})  ; 元信息，会按照原本的样子传给 alarm，可以在获取报警的时候看到这个信息
  "
  [m]
  (let [conn {:pool {}, :spec {:uri @alarm-url}}]
    (fn [ev]
      (let [f #(let [v (% m)] (if (fn? v) (v ev) v))
            tags (or (f :outstanding-tags) (all-tags ev))
            event? (let [v (f :event?)] (if (nil? v) false v))
            status (if event? :event (:state ev))
            alarmev {:version "alarm-V1"
                     :id (evid ev tags m)
                     :time (:time ev)
                     :level (f :level)
                     :status (string/upper-case (name status))
                     :endpoint (:host ev)
                     :metric (:service ev)
                     :tags (select-keys ev tags)
                     :note (f :note)
                     :description (:description ev)
                     :expected (or (f :expected) "-")
                     :actual (:metric ev)
                     :groups (f :groups)
                     :meta (:meta ev)}]
        (io
          (car/wcar conn
            (car/rpush (str "satori-events:" (f :level)) (json/write-str alarmev))))))))

(defn- cond-rewrite
  "重写条件， '(> 3.0) --> '#(> (:metric %) 3.0)"
  [& conds]
  (map (fn [l] `#(~(first l) (:metric %) ~@(rest l))) conds))

(defn judge-gapped*
  "判定事件的状态。rising 是 OK -> PROBLEM 的条件， falling 是 PROBLEM -> OK 的条件"
  [rising falling & children]
  (let [state (atom :ok)]
    (fn [ev]
      (let [current (if (= :ok @state)
                      (if (rising ev) :problem :ok)
                      (if (falling ev) :ok :problem))]
        (reset! state current)
        (call-rescue (assoc ev :state current) children)))))

(defmacro judge-gapped
  "参见 judge-gapped*，这里 c 是形如 (> 1.0) 的 form"
  [rising falling & children]
  `(judge-gapped* ~@(cond-rewrite rising falling) ~@children))

(defn judge*
  "判定事件的状态。c 是接受事件作为参数的函数。c 返回真值代表有问题。"
  [c & children]
  (fn [ev]
    (call-rescue (assoc ev :state (if (c ev) :problem :ok)) children)))

(defmacro judge
  "参见 judge*，这里 c 是形如 (> 1.0) 的 form"
  [c & children]
  `(judge* ~@(cond-rewrite c) ~@children))

; For backward compatibility
(def set-state* judge*)
(def set-state-gapped* judge-gapped*)
(defmacro set-state [& args] `(judge ~@args))
(defmacro set-state-gapped [& args] `(judge-gapped ~@args))


(defn alarm-every
  "如果是 PROBLEM 状态，每 dt 时间通知一次。变回 OK 的时候通知一次。
   unit 是 dt 的时间单位，可以是 :sec :secs :min :mins :hour :hours （单复形式无区别）"
  [dt unit & children]
  (let [state (atom :ok)
        last-notify (atom 0)
        dt (* ({:sec 1 :secs 1 :second 1 :seconds 1
                :min 60 :mins 60 :minute 60 :minutes 60
                :hour 3600 :hours 3600} unit) dt)]
    (fn [ev]
      (cond
        ;; rising
        (= :problem (:state ev))
        (do (reset! state :problem)
            (when (> (- (unix-time) @last-notify) dt)
              (reset! last-notify (unix-time))
              (call-rescue ev children)))

        ;; falling
        (and (= :ok (:state ev))
             (= :problem @state))
        (do (reset! state :ok)
            (reset! last-notify 0)
            (call-rescue ev children))))))

; For backward compatibility
(defn should-alarm-every
  [dt & children]
  (apply alarm-every (concat [dt :secs] children)))

(defn quiet?
  "接收数组格式的 rules 作为静默规则， 如 [['backup','* * 1-9 * * ?']]
   backup 是个正则（也可以为 string), 用来匹配 host(endpoint)
   '* * 1-9 * * ?' 为 cron 的表达式(local timezone)，用来表示生效的时间范围"
  [rule & children]
  ;; 只能处理单个 event
  (fn [evt]
    (try
      (let [matched-rule (-> (filter
                              (fn [r] (or (= (:host evt) (first r))
                                          (not-empty
                                            (re-matches (re-pattern (first r)) (:host evt))))) rule)
                              (first)
                              (second))
            time (java.util.Date. (* (:time evt) 1000))]
        (if matched-rule
          ;; http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html
          (let [cron-exp (CronExpression. matched-rule)]
            ;; fix timezone later
            (.setTimeZone cron-exp (java.util.TimeZone/getTimeZone "Asia/Shanghai"))
            (when-not (.isSatisfiedBy cron-exp time)
              (call-rescue evt children)))
          (call-rescue evt children)))
      (catch Exception e
        (info e)))))

(defn feed-dog
  "喂狗。如果 ttl 之内没有再次喂狗，就会触发 watchdog 报警，配合 watchdog 流使用。"
  ([ttl]
    (feed-dog ttl []))

  ([ttl outstanding-tags]
    (smap #(event {:service (evid % outstanding-tags {}), :metric %})
      (with {:host ".satori.watchdog.bark", :ttl ttl} (index)))))


(defn watchdog
  "看门狗（Open-Falcon 的 nodata 功能）的流。这个流不能接在 where 后面，必须看到所有的事件。
   会把所有过期的事件传递到下面，自己接 where 过滤。"
  [& children]
  (let [state (atom :ok)]
    (sdo
      (where (and (host ".satori.watchdog.bark")
                  (expired? event))
        (smap #(into % {:state nil, :time (unix-time)})
          (index)
          (smap #(into % {:host ".satori.watchdog.calm"
                          :ttl (* (:ttl %) 2)})
            (index)))

        (smap :metric
          (with {:state :problem}
            (apply sdo children))))

      (where (and (host ".satori.watchdog.calm")
                  (expired? event))
        (smap :metric
          (with {:state :ok}
            (apply sdo children)))))))
