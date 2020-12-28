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

(def watcher (fw/watch "test" callback {:delay-ms 250 :recursive true}))

(Thread/sleep 200)
(sh "touch" *file*)
(Thread/sleep 3000)

(prn :events @events)

(fw/unwatch watcher)

(println :done)
