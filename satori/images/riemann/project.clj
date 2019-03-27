(defproject riemann "0.3.1-satori"
  :description "Satori flavored Riemann"
  :url "http://example.com/FIXME"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :plugins [[io.aviso/pretty "0.1.37"]]
  :middleware [io.aviso.lein-pretty/inject]
  :dependencies [[org.clojure/clojure "1.10.0"]
                 [org.clojure/data.json "0.2.6"]
                 [com.taoensso/carmine "2.19.1"]
                 [io.aviso/pretty "0.1.37"]
                 [riemann "0.3.1"]])
