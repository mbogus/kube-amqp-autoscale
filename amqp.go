// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/streadway/amqp"
)

var (
	queueCountSuccesses = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "queue_count_successes_total",
		Help:      "Number of successful queue count retrievals.",
	})
	queueCountFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "queue_count_failures_total",
		Help:      "Number of failed queue count retrievals.",
	})
	currentQueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "current_queue_size",
			Help:      "Last count retrieved for a queue.",
		},
		[]string{"queue"},
	)
	metricSaveFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "metric_save_failures_total",
		Help:      "Number of times saving metrics failed.",
	})
)

type saveStat func(int) error

func monitorQueue(uri string, names []string, interval int, f saveStat, quit <-chan struct{}) {
	for {
		select {
		case <-quit:
			return
		case <-time.After(time.Duration(interval) * time.Second):
			totalMsgs := 0
			errored := false
			for _, name := range names {
				msgs, err := getQueueLength(uri, name)
				if err != nil {
					queueCountFailures.Inc()
					log.Printf("Failed to get queue length for queue %s: %v", name, err)
					errored = true
				} else {
					totalMsgs += msgs
					queueCountSuccesses.Inc()
					currentQueueSize.WithLabelValues(name).Set(float64(msgs))
				}
			}
			// Only save metrics if both counts succeeded.
			if errored == false {
				err := f(totalMsgs)
				if err != nil {
					metricSaveFailures.Inc()
					log.Printf("Error saving metrics: %v", err)
				}
			}
		}
	}
}

func getQueueLength(uri, name string) (int, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()
	q, err := ch.QueueInspect(name)
	if err != nil {
		return 0, err
	}
	return q.Messages, nil
}
