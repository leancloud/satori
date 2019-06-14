(ns agent-plugin
  (:require [clojure.set :as cset]
            [clojure.string :as string]
            [clojure.stacktrace :as st]
            [clojure.data.json :as json]
            [clojure.tools.logging :refer [info error]]
            [clojure.java.shell :refer [sh *sh-dir*]]
            [clojure.core.async :refer [chan go-loop <! >! <!! >!! mix admix timeout alts!]]

            [riemann.service :as service]
            [riemann.config :refer :all]
            [riemann.common :refer [event]]
            [riemann.test :refer [io]]
            [riemann.bin :refer [reload!]]
            [taoensso.carmine :as car])

  (:import (java.util.concurrent.locks LockSupport)))


(defmacro def- [& forms]
  `(def ^{:private true} ~@forms))

(def- state (atom {}))
#_({"dns1" {:dirs #{"a" "b"}
            :plugins #{{:_metric "a", :_step 30 :a 1}
                       {:_metric "b", :_step 30 :b 2}}}})

(def- masters (atom nil))

(defn- debounce [in ms]
  (let [out (chan)]
    (go-loop [last-val nil]
      (let [val (if (nil? last-val) (<! in) last-val)
            timer (timeout ms)
            [new-val ch] (alts! [in timer])]
        (condp = ch
          timer (do (>! out val) (recur nil))
          in (if new-val (recur new-val)))))
    out))


(def- update-channel (chan))
(def- update-channel-mix (mix update-channel))
(def- debouncers (atom {}))
(def- master-informer
  (go-loop []
    (let [v (json/write-str (<! update-channel))]
      (doseq [m @masters]
        (car/wcar m
          (car/publish "satori:master-state" v))))
    (recur)))


(defn- inform-master
  [state]
  (let [k (str (:type state) ":" (:hostname state))
        ch (or (get @debouncers k)
               (let [ch' (chan)]
                 (admix update-channel-mix (debounce ch' 3000))
                 (swap! debouncers (fn [v v'] (merge v' v)) {k ch'})
                 (get @debouncers k)))]
    (>!! ch state)))

(defn- update-plugin-dirs!
  [host dirs]
  (inform-master {:type "plugin-dir",
                  :hostname host,
                  :dirs dirs}))

(defn- update-plugins!
  [host plugins]
  (inform-master {:type "plugin",
                  :hostname host,
                  :plugins plugins}))

(defn watch-for-master-restart!
  [uri]
  (service! (service/thread-service
    ::watch-for-master-restart uri (fn [_]
    (as-> nil listener
      (car/with-new-pubsub-listener {:uri uri}
        {"satori:component-started" (fn [[_ _ component]]
          (when (= component "master")
            (info "Master restart detected, reloading...")
            (future (reload!))))}
        (car/subscribe "satori:component-started"))
      (try
        (LockSupport/park)
        (finally
          (car/close-listener listener))))))))

(defn set-master-redis!
  "指定 master 的 redis 地址，可以指定多个"
  [& uris]
  (when (not= @masters nil)
    (throw (Exception. "Masters already set")))

  (doseq [uri uris]
    (watch-for-master-restart! uri))

  (reset! masters
    (for [uri uris] {:pool {}, :spec {:uri uri}})))

(defn set-plugin-version!
  "指定插件的版本，需要是完整的 git commit hash"
  [v]
  (inform-master {:type "plugin-version", :version v}))

(defn watch-for-update!
  [path old]
  (service! (service/thread-service
    ::watch-for-update [path old] (fn [_]
    (binding [*sh-dir* path]
      (Thread/sleep 5000)
      (let [r (sh "git" "rev-parse" "HEAD")]
        (if (not= 0 (:exit r))
          (do
            (error "Can't determine plugin version in watch-for-update:\n" (:err r))
            (Thread/sleep 30000))

          (let [commit (string/trim (:out r))]
            (when (not= old commit)
              (info "Rules repo updated, reloading...")
              (sh "git" "reset" "--hard")
              (sh "git" "clean" "-f" "-d")
              (future (let [rst (reload!)]
                (if (= rst :reloaded)
                  (do (info "Report reload success")
                  (reinject (event {:service ".satori.riemann.newconf"
                                    :host "Satori"
                                    :metric 1
                                    :description commit})))
                  (let [ex (with-out-str (st/print-stack-trace rst))]
                    (info "Report reload failure")
                    (reinject (event {:service ".satori.riemann.reload-failed"
                                      :host "Satori"
                                      :metric 1
                                      :description ex}))))))
              (Thread/sleep 120000))))))))))

(defn set-plugin-repo!
  "指定插件的版本，需要指定 git 仓库的地址，需要是本地地址"
  [path]
  (io
    (binding [*sh-dir* path]
      (let [r (sh "git" "rev-parse" "HEAD")]
        (when (not= 0 (:exit r))
          (throw (Exception. (str "Can't determine plugin version:\n" (:err r)))))
        (let [commit (string/trim (:out r))]
          (set-plugin-version! commit)
          (watch-for-update! path commit))))))


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
            (update-plugin-dirs! h (:dirs (get rst h)))))))))

(defn plugin
  "为机器指定一个插件。 metric 就是类似于 net.port.listen 的，也就是 riemann 中的 service。
   step 是收集的间隔。args 是个 map，会传给插件当做参数。"
  [metric step args]
  (let [m (conj {:_metric metric :_step step} args)]
    (fn [ev]
      (let [h (:host ev)]
        (when-not ((or (:plugins (get @state h)) #{}) m)
          (as-> nil rst
            (swap! state (fn [state']
              (update-in state' [h :plugins] #(conj (or % #{}) m))))
            (update-plugins! h (:plugins (get rst h)))))))))


(def plugin-metric plugin)  ; for backward compatibility
