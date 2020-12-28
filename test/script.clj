#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]))

(pods/load-pod "./main")

(require '[pod.babashka.filewatcher :as fw])

(def watcher (fw/watch "." (fn [event] (prn event)) {:delay-ms 2500 :recursive true}))

(prn :watcher watcher)

@(promise)

