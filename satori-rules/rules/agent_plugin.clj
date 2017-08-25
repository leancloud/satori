(ns agent-plugin
  (:require [taoensso.carmine :as car]
            [clojure.set :as cset]
            [clojure.string :as string]
            [clojure.data.json :as json])
  (:use riemann.streams
        riemann.time
        clojure.java.shell))

(defmacro def- [& forms]
  `(def ^{:private true} ~@forms))

(def- state (atom {}))
#_({"dns1" {:dirs #{"a" "b"}
            :metrics #{{:_metric "a", :_step 30 :a 1}
                       {:_metric "b", :_step 30 :b 2}}}})

(def- cooled (atom false))
(def- cooling-time 70)
(def- satori-masters (atom [{:pool {}, :sepc {:uri "redis://localhost:6379/0"}}]))

(future
  (Thread/sleep (* cooling-time 1000))
  (reset! cooled true)
  (doseq [[host s] @state]
    (let [dirs    (json/write-str {:type "plugin-dir",    :hostname host, :dirs (:dirs s)})
          metrics (json/write-str {:type "plugin-metric", :hostname host, :metrics (:metrics s)})]
      (doseq [m @satori-masters]
        (car/wcar m
          (car/publish "satori:master-state" dirs)
          (car/publish "satori:master-state" metrics))))))


(defn defmaster
  "指定 master 的 redis 地址，可以指定多个"
  [& masters]
  (reset! satori-masters
    (for [uri masters] {:pool {}, :spec {:uri uri}})))

(defn set-plugin-version
  "指定插件的版本，需要是完整的 git commit hash"
  [v]
  (let [s (json/write-str {:type "plugin-version", :version v})]
    (doseq [m @satori-masters]
      (car/wcar m (car/publish "satori:master-state" s)))))


(defn set-plugin-repo
  "指定插件的版本，需要指定 git 仓库的地址，需要是本地地址"
  [p]
  (binding [*sh-dir* p]
    (let [r (sh "git" "rev-parse" "HEAD")]
      (when (not (= 0 (:exit r)))
        (throw (Exception. (str "Can't determine plugin version:\n" (:err r)))))
      (set-plugin-version (string/trim (:out r))))))


(defn plugin-dir
  "为机器指定插件目录。所有流过这个 stream 的 event 中的 host 都会执行 dirs 里面的插件。"
  [& dirs]
  (fn [ev]
    (when-not @cooled
      (let [dirs (set (map #(if (fn? %) (% ev) %) dirs))
            h (:host ev)]
        (when-not (cset/subset? dirs (get-in @state [h :dirs]))
          (swap! state (fn [state']
            (update-in state' [h :dirs] #(cset/union % dirs)))))))))

(defn plugin-metric
  "为机器指定一个 metric 插件。 metric 就是类似于 net.port.listen 的，也就是 riemann 中的 service。
   step 是收集的间隔。args 是个 map，会传给插件当做参数。"
  [metric step args]
  (fn [ev]
    (when-not @cooled
      (let [m (into {:_metric metric :_step step}
                (for [[k v] (seq args)]
                  [k (if (fn? v) (v ev) v)]))
            h (:host ev)]
        (when-not ((or (get-in @state [h :metrics]) #{}) m)
          (swap! state (fn [state']
            (update-in state' [h :metrics] #(conj (or % #{}) m)))))))))
