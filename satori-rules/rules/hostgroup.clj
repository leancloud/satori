(ns hostgroup)

(defmacro is? [re]
  `(re-matches ~re ~'host))

(defn host->group [ev]
  (let [host (:host ev)]
    (cond
      (is? #"^docker\d+") [:operation :lean-engine]
      (is? #"^api\d+")    [:operation :api]
      (is? #"^push\d+")   [:operation :push]
      (is? #"^stats\d+")  [:operation :stats]
      :else               [:operation])))
