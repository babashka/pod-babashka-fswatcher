#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.shell :refer [sh]]
            [clojure.test :as t :refer [deftest is testing]]))

(prn (pods/load-pod "./pod-babashka-fswatcher"))

(require '[pod.babashka.fswatcher :as fw])

(def events (atom []))

(def callback
  (fn [event]
    ;; (prn :event event)
    (swap! events conj event)))

(def watcher (fw/watch "test" callback {:delay-ms 250 :recursive true}))

(Thread/sleep 200)
(sh "touch" *file*)
(Thread/sleep 1000)

(prn :events @events)

(fw/unwatch watcher)
(fw/unwatch watcher) ;; idempotency

(def ev1 @events)

(sh "touch" *file*)
(Thread/sleep 1000)

(def ev2 @events)

(deftest events-test
  (is (pos? (count ev1)))
  (is (contains? (set (map :path ev1)) "test/script.clj"))
  (testing "No new events after unwatch"
    (is (= (count ev1) (count ev2)))))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
