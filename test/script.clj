#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.shell :refer [sh]]))

(prn (pods/load-pod "./main"))

(require '[pod.babashka.filewatcher :as fw])

(def events (atom []))

(def callback
  (fn [event]
    (prn :event event)
    (swap! events conj event)))

(def watcher (fw/watch "test" callback {:delay-ms 2500 :recursive true}))

(Thread/sleep 1000)
(sh "touch" *file*)
(Thread/sleep 1000)

(prn :events @events)

(fw/unwatch watcher)

(println :done)
