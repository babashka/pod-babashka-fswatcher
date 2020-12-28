#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]))

(pods/load-pod "./main")

(require '[pod.babashka.filewatcher :as fw])

(fw/watch "." (fn [event] (prn event)) {:delay-ms 50 :recursive true})

@(promise)

