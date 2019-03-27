(ns agent-plugin
  (:require [taoensso.carmine :as car]
            [clojure.set :as cset]
            [clojure.string :as string]
            [clojure.data.json :as json])
  (:use riemann.streams
        riemann.time
        [riemann.test :only [io]]
        clojure.java.shell))

(defmacro def- [& forms]
  `(def ^{:private true} ~@forms))

(def- state (atom {}))
#_({"dns1" {:dirs #{"a" "b"}
            :metrics #{{:_metric "a", :_step 30 :a 1}
                       {:_metric "b", :_step 30 :b 2}}}})

(def- satori-masters (atom [{:pool {}, :spec {:uri "redis://localhost:6379/0"}}]))

(defn- update-plugin-dirs
  [host dirs]
  (let [dirs (json/write-str {:type "plugin-dir",
                              :hostname host,
                              :dirs dirs})]
    (io
      (doseq [m @satori-masters]
        (car/wcar m
          (car/publish "satori:master-state" dirs))))))


(defn- update-plugins
  [host metrics]
  (let [metrics (json/write-str {:type "plugin",
                                 :hostname host,
                                 :metrics metrics})]
    (io
      (doseq [m @satori-masters]
        (car/wcar m
          (car/publish "satori:master-state" metrics))))))

(defn defmaster
  "指定 master 的 redis 地址，可以指定多个"
  [& masters]
  (reset! satori-masters
    (for [uri masters] {:pool {}, :spec {:uri uri}})))


(defn set-plugin-version
  "指定插件的版本，需要是完整的 git commit hash"
  [v]
  (io
    (let [s (json/write-str {:type "plugin-version", :version v})]
      (doseq [m @satori-masters]
        (car/wcar m (car/publish "satori:master-state" s))))))

(defn set-plugin-repo
  "指定插件的版本，需要指定 git 仓库的地址，需要是本地地址"
  [p]
  (io
    (binding [*sh-dir* p]
      (let [r (sh "git" "rev-parse" "HEAD")]
        (when (not (= 0 (:exit r)))
          (throw (Exception. (str "Can't determine plugin version:\n" (:err r)))))
        (set-plugin-version (string/trim (:out r)))))))

(defn plugin-dir
  "为机器指定插件目录。所有流过这个 stream 的 event 中的 host 都会执行 dirs 里面的插件。"
  [& dirs]
  (let [dirs (set dirs)]
    (fn [ev]
      (let [h (:host ev)]
        (when-not (cset/subset? dirs (:dirs (get @state h)))
          (as-> nil rst
            (swap! state (fn [state']
              (update-in state' [h :dirs] #(cset/union % dirs))))
            (update-plugin-dirs h (:dirs (get rst h)))))))))

(defn plugin
  "为机器指定一个插件。 metric 就是类似于 net.port.listen 的，也就是 riemann 中的 service。
   step 是收集的间隔。args 是个 map，会传给插件当做参数。"
  [metric step args]
  (let [m (conj {:_metric metric :_step step} args)]
    (fn [ev]
      (let [h (:host ev)]
        (when-not ((or (:metrics (get @state h)) #{}) m)
          (as-> nil rst
            (swap! state (fn [state']
              (update-in state' [h :metrics] #(conj (or % #{}) m))))
            (update-plugins h (:metrics (get rst h)))))))))
