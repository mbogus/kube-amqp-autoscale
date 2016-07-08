// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

type saveStat func(int) error

func monitorQueue(uri, name string, interval int, f saveStat, quit <-chan struct{}) {
	for {
		select {
		case <-quit:
			return
		case <-time.After(time.Duration(interval) * time.Second):
			msgs, err := getQueueLength(uri, name)
			if err == nil {
				err = f(msgs)
			}
			if err != nil {
				log.Printf("Error saving metrics: %v", err)
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
