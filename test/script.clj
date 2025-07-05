#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
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
  (is (= 1 (count ev1)))
  (is (= (:path (first ev1)) "test/script.clj"))
  (testing "No new events after unwatch"
    (is (= (count ev1) (count ev2)))))

(deftest dedup-test
  (reset! events [])
  (let [watcher (fw/watch "test" #(swap! events conj %) {:delay-ms 50 :recursive true})]
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
  (reset! events [])
  (let [watcher (fw/watch "test" #(swap! events conj %) {:delay-ms 50 :recursive true})]
    (sh "touch" *file*)
    (Thread/sleep 51)
    (sh "touch" *file*)
    (Thread/sleep 60)
    (prn :events-dedup-outside-interval @events)
    (testing "events outside of dedup interval come through."
      (is (= 2 (count @events))))
    (fw/unwatch watcher)))

(deftest recursive-dedup-test
  (let [ev (atom [])
        file-name "test/dir/anotherdir/bla.txt"
        _ (clojure.java.io/make-parents file-name)
        watcher (fw/watch "test" #(swap! ev conj %) {:delay-ms 50 :recursive true})]
    (spit file-name "whatever")
    (Thread/sleep 60)
    (prn :events-recursive-dedup @ev)
    (testing "dedup recursive works"
      (is (= @ev [{:type :write, :path "test/dir/anotherdir/bla.txt"}]))
      (fw/unwatch watcher))))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
