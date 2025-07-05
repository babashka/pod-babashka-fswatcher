#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.java.shell :refer [sh]]
            [clojure.test :as t :refer [deftest is testing]]))

(prn (pods/load-pod "./pod-babashka-fswatcher"))

(require '[pod.babashka.fswatcher :as fw])

 ;; idempotency

(deftest events-test
  (let [events (atom [])
        callback
        (fn [event]
          ;; (prn :event event)
          (swap! events conj event))
        watcher (fw/watch "test" callback {:delay-ms 250 :recursive true})]
    (Thread/sleep 200)
    (sh "touch" *file*)
    (Thread/sleep 500)
    (let [ev1 @events]
      (fw/unwatch watcher)
      (fw/unwatch watcher)
      (sh "touch" *file*)
      (Thread/sleep 1000)
      (let [ev2 @events]
        (is (= 1 (count ev1)))
        (is (= "test/script.clj" (:path (first ev1))))
        (testing "No new events after unwatch"
          (is (= (count ev1) (count ev2))))))))

(deftest dedup-test
  (let [events (atom [])
        watcher (fw/watch "test" #(swap! events conj %) {:delay-ms 50 :recursive true})]
    (sh "touch" *file*)
    (Thread/sleep 5)
    (sh "touch" *file*)
    (Thread/sleep 5)
    (sh "touch" *file*)
    ;;wait for timer to end
    (Thread/sleep 60)
    (prn :events-dedup @events)
    (testing "tests that the events that happened inside the interval were deduped."
      (is (= 1 (count @events))))
    (fw/unwatch watcher)))

(deftest dedup-outside-interval-test
  (let [events (atom [])
        watcher (fw/watch "test" #(swap! events conj %) {:delay-ms 50 :recursive true})]
    (sh "touch" *file*)
    (Thread/sleep 51)
    (sh "touch" *file*)
    (Thread/sleep 60)
    (prn :events-dedup-outside-interval @events)
    (testing "events outside of dedup interval come through."
      (is (= 2 (count @events))))
    (fw/unwatch watcher)))

(deftest recursive-dedup-test
  (let [ ev (atom [])
        file-name "test/dir/anotherdir/bla.txt"
        _ (clojure.java.io/make-parents file-name)
        watcher (fw/watch "test" #(swap! ev conj %) {:delay-ms 50 :recursive true})]
    (spit file-name "whatever")
    (Thread/sleep 100)
    (prn :events-recursive-dedup @ev)
    (testing "dedup recursive works"
      (is (= "test/dir/anotherdir/bla.txt" (-> @ev first :path)))
      (fw/unwatch watcher))))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
