(ns lib
  (:require [clojure.set :as cset]
            [clojure.string :as string]
            [clojure.walk :refer [postwalk macroexpand-all]]

            [riemann.config :refer :all]
            [riemann.streams :refer :all]
            [riemann.test :refer [tests io tap]])

  (:import riemann.codec.Event))

(defmacro ->waterfall
  "接受一组 form，将后面的 form 插入到前面 form 的最后。
   如果你的规则从上到下只有一条分支，用这个可以将缩进压平，变的好看些。"
  [& forms] `(->> ~@(reverse forms)))

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
      (let [m (mapv :metric evts), v (f m)
            s (string/join "\n" (map #(str (or (:aggregate-desc-key %) (:host %)) "=" (:metric %)) evts))
            s (if (> (count s) 3000) (subs s 0 3000) s)
            aggregated (into (last evts) {:metric v, :description s})]
        (call-rescue aggregated children)))))

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
  (let [m (filter pos? m)
        r (last m)]
    (if r
      (->> m
           (map #(/ (- r %) (min r %)))
           (reduce #(if (|>| %1 %2) %1 %2) 0))
      0.0)))

(defn avgpdiff
  "计算最后一个点相比于之前的点的平均值的变化率"
  [& m]
  (let [r (last m)
        c (count m)]
    (if (not= r 0)
      (let [avg (/ (apply + (- r) m) (- c 1))]
        (/ (- r avg) (min avg r)))
      0)))


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
      (judge (> 0.5)
        (runs :state 5
          (alarm-every 2 :min
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
    (let [s (lib/slot-window :service {:foo "bar" :baz "quux"} (tap :slot-window))
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
    (let [s (lib/slot-window :service {:ev1 "metric.ev1" :ev2 "metric.ev2"}
              (lib/slot-coalesce {:service "metric.final"
                                    :metric [ev1 ev2
                                             (:metric ev:ev1) (:metric ev:ev2)
                                             (:metric (:ev1 event)) (:metric (:ev2 event))]}
                (tap :slot-coalesce-test)))
          rst (inject! [s] [{:host "meh" :service "metric.ev1" :metric 10},
                            {:host "meh" :service "metric.ev2" :metric 20}])]
      (is (= [{:host "meh" :service "metric.final" :metric [10 20 10 20 10 20]}]
             (:slot-coalesce-test rst)))))

  (deftest maxpdiff-test
    (is (= (lib/maxpdiff 1.0 1.0 1.0 1.0 1.0 2.0) 1.0))
    (is (= (lib/maxpdiff 2.0 1.0 1.0 1.0 1.0 1.0) -1.0))
    (is (= (lib/maxpdiff 0.0 0.0 0.0 0.0 1.0 2.0) 1.0))
    (is (= (lib/maxpdiff 0.0 0.0 0.0 0.0 0.0 0.0) 0.0)))

  (deftest aggregate-test
    (let [s (lib/aggregate + (tap :aggregate-test))
          rst (inject! [s] [[{:host "meh" :service "bar" :metric 10},
                            {:host "meh" :service "bar" :metric 80}]])]
      (is (= 90 (get-in rst [:aggregate-test 0 :metric]))))))
