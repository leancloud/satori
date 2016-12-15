(ns alarm
  (:require [taoensso.carmine :as car]
            [clojure.string :as string]
            [clojure.set :as cset]
            [clojure.data.json :as json])
  (:use riemann.streams
        riemann.config
        [riemann.common :only [event]]
        [riemann.time :only [unix-time]])
  (:import riemann.codec.Event))

(def the-index (index))

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

(defn defalarm [url]
  (reset! alarm-url url))

(defn !
  "
  创建一个发送报警的流，流经这个流的事件都会被发到 alarm 产生报警。
  接受一个 map 做参数，map 中需要可以指定如下的参数

  (! {:note \"报警标题，标题对于一个特定的报警是不能变的（不要把报警的数据编码在这里面）\"
      :level 1  ;报警级别, 0最高，6最小。报警级别影响报警方式。
      :event? false  ; 可选，是不是事件（而不是状态）。默认 false。如果是事件的话，只会发报警，不会记录状态（alarm插件里看不到）。
      :expected 233  ; 可选，期望值，暂时没用到)
      :outstanding-tags [:region :mount]  ; 可选，相关的tag，写在这里的 tag 会用于区分不同的事件，以及显示在报警内容中, 不填的话默认是所有的tag
      :groups [:operation]})  ; groups 是在规则仓库的 alarm 配置里管理的)
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
                     :groups (f :groups)}]
        (car/wcar conn
          (car/rpush (str "satori-events:" (f :level)) (json/write-str alarmev)))))))

(defn- cond-rewrite
  "重写条件， '(> 3.0) --> '#(> (:metric %) 3.0)"
  [& conds]
  (map (fn [l] `#(~(first l) (:metric %) ~@(rest l))) conds))

(defn set-state-gapped*
  "设置事件的状态。rising 是 OK -> PROBLEM 的条件， falling 是 PROBLEM -> OK 的条件"
  [rising falling & children]
  (let [state (atom :ok)]
    (fn [ev]
      (let [current (if (= :ok @state)
                      (if (rising ev) :problem :ok)
                      (if (falling ev) :ok :problem))]
        (reset! state current)
        (call-rescue (assoc ev :state current) children)))))

(defmacro set-state-gapped
  "参见 set-state-gapped*，这里 c 是形如 (> 1.0) 的 form"
  [rising falling & children]
  `(set-state-gapped* ~@(cond-rewrite rising falling) ~@children))

(defn set-state*
  "设置事件的状态。c 是接受事件作为参数的函数。c 返回真值代表有问题。"
  [c & children]
  (fn [ev]
    (call-rescue (assoc ev :state (if (c ev) :problem :ok)) children)))

(defmacro set-state
  "参见 set-state*，这里 c 是形如 (> 1.0) 的 form"
  [c & children]
  `(set-state* ~@(cond-rewrite c) ~@children))


(defn should-alarm-every
  "如果是 PROBLEM 状态，每 dt 秒通知一次。变回 OK 的时候通知一次。"
  [dt & children]
  (let [state (atom :ok)
        last-notify (atom 0)]
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

(defn aggregate*
  "接受事件的数组，变换成一个事件。
   f 函数接受 metric 数组作为参数，返回一个 aggregate 过的 metric，
   最后将 aggregate 过的 metric 放在最后一个事件中，向下传递这个事件。
   常接在 (moving|fixed)-(time|event)-window 流后面。"
  [f & children]
  (fn [evts]
    (when-not (sequential? evts)
      (throw (Exception. "aggregate must accept list of events")))

    (when (> (count evts) 0)
      (let [m (mapv :metric evts), v (f m)]
        (call-rescue
          (into (last evts) {:metric v, :description (string/join "\n" (map #(str (:host %) "=" (:metric %)) evts))})
          children)))))

(defn aggregate [f & children]
  (apply aggregate* (fn [m] (apply f m)) children))

(defn copy
  "将事件的一个 field 复制到另一个 field。通常接在 aggregate 后面，用于修正 host。
   (copy :region :host (...))"
  [from to & children]
  (apply smap #(assoc % to (from %)) children))


(defn |>| [& args] (apply > (map #(Math/abs %) args)))
(defn |<| [& args] (apply < (map #(Math/abs %) args)))

(defn maxpdiff
  "计算最大变化率，与 aggregate 搭配使用(MAX Percentage DIFFerence)"
  [& m]
  (let [r (last m)]
    (->> (map #(/ (- r %) r) m)
         (reduce (fn [v v'] (if (> (Math/abs v') (Math/abs v)) v' v))))))


(defn ->difference
  "接受事件的数组，变换成另外一个事件数组。
   新的事件数组中，每一个事件的 metric 是之前相邻两个事件 metric 的差。
   如果你的事件是一个一直增长的计数器，那么用这个流可以将它变成每次实际增长的值。"
  [& children]
  (apply smap (fn [l]
    (map
      (fn [[ev' ev]] (assoc ev' :metric (- (:metric ev') (:metric ev))))
      (map vector (rest l) l)))
    children))


(defn feed-dog
  "喂狗。如果 ttl 之内没有再次喂狗，就会触发 watchdog 报警，配合 watchdog 流使用。"
  ([ttl]
    (feed-dog ttl []))

  ([ttl outstanding-tags]
    (smap #(event {:service (evid % outstanding-tags {}), :metric %})
      (with {:host ".satori.watchdog.bark", :ttl ttl} the-index))))


(defn watchdog
  "看门狗（Open-Falcon 的 nodata 功能）的流。这个流不能接在 where 后面，必须看到所有的事件。
   会把所有过期的事件传递到下面，自己接 where 过滤。"
  [& children]
  (let [state (atom :ok)]
    (sdo
      (where (and (host ".satori.watchdog.bark")
                  (expired? event))
        (smap #(into % {:state nil, :time (unix-time)})
          the-index
          (smap #(into % {:host ".satori.watchdog.calm"
                          :ttl (* (:ttl %) 2)})
            the-index))

        (smap :metric
          (with {:state :problem}
            (apply sdo children))))

      (where (and (host ".satori.watchdog.calm")
                  (expired? event))
        (smap :metric
          (with {:state :ok}
            (apply sdo children)))))))
