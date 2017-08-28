(ns alarm
  (:require [taoensso.carmine :as car]
            [clojure.string :as string]
            [clojure.set :as cset]
            [clojure.data.json :as json])
  (:use riemann.streams
        riemann.config
        [riemann.test :only [tests io tap]]
        [clojure.walk :only [postwalk macroexpand-all]]
        [riemann.common :only [event]]
        [riemann.time :only [unix-time]])
  (:import riemann.codec.Event))

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

  (! {:note string ; 报警标题，标题对于一个特定的报警是不能变的（不要把报警的数据编码在这里面）
      :level 1  ;报警级别, 0最高，6最小。报警级别影响报警方式。
      :event? false  ; 可选，是不是事件（而不是状态）。默认 false。如果是事件的话，只会发报警，不会记录状态（alarm插件里看不到）。
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
        (io
          (car/wcar conn
            (car/rpush (str "satori-events:" (f :level)) (json/write-str alarmev))))))))

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


(defn |>| [& args] (apply > (map #(Math/abs %) args)))
(defn |<| [& args] (apply < (map #(Math/abs %) args)))

(defn maxpdiff
  "计算列表中最后一个点相对之前的点的最大变化率(MAX Percentage DIFFerence)，
   与 aggregate 搭配使用。计算变化率时总是使用两个点中的最小值做分母，
   所以由1变到2的变化率是 1.0, 由2变到1的变化率是 -1.0 （而不是 -0.5)
   "
  [& m]
  (let [r (last m)]
    (->> (map #(/ (- r %) (min r %)) m)
         (reduce (fn [v v'] (if (> (Math/abs v') (Math/abs v)) v' v))))))

(defn avgpdiff
  "计算最后一个点相比于之前的点的平均值的变化率"
  [& m]
  (let [avg (/ (apply + (- (last m)) m) (- (count m) 1))]
    (/ (- (last m) avg) (min avg (last m)))))


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

(defn group-window
  "将事件分组后向下传递，类似 fixed-time-window，但不使用时间切割，
  而是通过 (group-fn event) 的值进行切割。(group-fn event) 的值会被记录下来，
  每一次出现重复值的时候，会将当前缓存住的事件数组向下传递。

  比如你有一组同质的机器，跑了相同的服务，但是机器名不一样，可以通过
  (group-window :host
    ....)
  将事件分组后处理（e.g. 对单台的容量求和获得总体容量）
  e.g.: 一个事件流中的事件先后到达，其中 :host 的值如下
      a b c d b a c a b
  那么会被这个流分成
    [a b c d] [b a c] [a b]
  分成 3 次向下游传递
  "
  [group-fn & children]
  (let [buffer (ref [])
        group-keys (ref #{})]
    (fn stream [event]
      (let [evkey (group-fn event)]
        (as-> nil rst
          (dosync
            (if (@group-keys evkey)
              (let [events @buffer]
                (ref-set buffer [event])
                (ref-set group-keys #{evkey})
                events)
              (do
                (alter buffer conj event)
                (alter group-keys conj evkey)
                nil)))
          (when rst
            (call-rescue rst children)))))))

(defn slot-window*
  [slot-fn fields children]
  (let [valid (set (vals fields))
        invert (cset/map-invert fields)
        current (ref {})
        remaining (ref (set (vals fields)))]
    (fn stream [event]
      (let [evkey (slot-fn event)]
        (as-> nil rst
          (when (valid evkey)
            (dosync
              (alter current assoc (invert evkey) event)
              (alter remaining disj evkey)
              (if (empty? @remaining)
                (let [r @current]
                  (ref-set current {})
                  (ref-set remaining valid)
                  r)
                nil)))
          (when rst
            (call-rescue rst children)))))))

(def ^{:private true, :dynamic true} *slot-window-slots*)

(defmacro slot-window
  "收集指定的几个事件并打包向下传递。事件的特征由 slot-fn 提取，并与 fields 中的
  的定义匹配，如果 fields 中的所有条件匹配的事件都收集到了，则打包向下传递并开始下一轮收集。
  与 group-window 相反，group-window 收集一组同质的事件，slot-window 用于收集一组异质的事件。
  当 slot-window 遇到重复的事件但是还没有满足向下传递的条件时，新的事件会替换掉缓存住的已有事件。

  常用于收集同一个资源不同的 metric 用于复杂的判定。

  比如有一个服务，同时收集了错误请求量和总量，希望按照错误数量在一定之上后按照错误率报警

  (slot-window :service {:error \"app.req.error\"
                         :count \"app.req.count\"}

    ; 此时会有形如 {:error {:service \"app.req.error\", ...},
    ;               :count {:service \"app.req.count\", ...}} 的事件传递下来

    ; 构造出想要的 event
    (slot-coalesce {:service \"app.req.error_rate\"
                    :metric (if (> error 100) (/ error count) -1)}
      (set-state (> 0.5)
        (runs :state 5
          (should-alarm-every 120
            (! {...
                ...}))))))
  "
  [slot-fn fields & children]
  (binding [*slot-window-slots* fields]
    `(slot-window* ~slot-fn ~fields ~(macroexpand-all (vec children)))))


(defn- ev-rewrite-slot
  [varname form fields]
  (let [vars (->> fields (keys) (map name) (map symbol) (set))
        evvars (->> fields
                    (keys)
                    (map #(vector (symbol (str "ev:" (name %))) (keyword %)))
                    (into {}))]
    (postwalk (fn [node]
      (if (symbol? node)
        (cond
          (vars node) `(:metric (~(keyword node) ~varname))
          (evvars node) (list (evvars node) varname)
          (= node 'event) varname
          :else node)
        node)) form)))

(defmacro slot-coalesce
  "对 slot-window 的结果进行计算，并构造出单一的 event。
  具体用法可以看 slot-window 的帮助

  ev': 构造出的新 event 模板。表达式中可以直接用如下的约定引用 slot 中的值：
    ; 假设: (slot-window :service {:some-counter1 \"app.some_counter\"} ...)
    some-counter1 ; :some-counter1 的 metric 值
    ev:some-counter1 ; :some-counter1 的整个 event
    event ; slot-window 整个传递下来的 {:some-counter1 ...}
  "
  [ev' & children]
  (if (bound? #'*slot-window-slots*)
    `(smap
      (fn [~'event]
        (conj (select-keys (first (vals ~'event)) [:host :time])
              ~(ev-rewrite-slot 'event ev' *slot-window-slots*)))
      ~@children)
    (throw (Exception. "Could not find slot-window stream!"))))


; ------------------------------------------------------------------
(tests
  (deftest slot-window-test
    (let [s (alarm/slot-window :service {:foo "bar" :baz "quux"} (tap :slot-window))
          rst (inject! [s] [{:host "meh" :service "bar" :metric 10},
                            {:host "meh" :service "quux" :metric 20},
                            {:host "meh" :service "bar" :metric 30},
                            {:host "meh" :service "irrelevant" :metric 35},
                            {:host "meh" :service "bar" :metric 40},
                            {:host "meh" :service "quux" :metric 50},
                            {:host "meh" :service "quux" :metric 60},
                            {:host "meh" :service "bar" :metric 70},
                            {:host "meh" :service "bar" :metric 80}])]
      (is (= [{:foo {:host "meh" :service "bar" :metric 10}
               :baz {:host "meh" :service "quux" :metric 20}},
              {:foo {:host "meh" :service "bar" :metric 40}
               :baz {:host "meh" :service "quux" :metric 50}},
              {:foo {:host "meh" :service "bar" :metric 70}
               :baz {:host "meh" :service "quux" :metric 60}}] (:slot-window rst)))))

  (deftest slot-coalesce-test
    (let [s (alarm/slot-window :service {:ev1 "metric.ev1" :ev2 "metric.ev2"}
              (alarm/slot-coalesce {:service "metric.final"
                                    :metric [ev1 ev2
                                             (:metric ev:ev1) (:metric ev:ev2)
                                             (:metric (:ev1 event)) (:metric (:ev2 event))]}
                (tap :slot-coalesce-test)))
          rst (inject! [s] [{:host "meh" :service "metric.ev1" :metric 10},
                            {:host "meh" :service "metric.ev2" :metric 20}])]
      (is (= [{:host "meh" :service "metric.final" :metric [10 20 10 20 10 20]}]
             (:slot-coalesce-test rst))))))
