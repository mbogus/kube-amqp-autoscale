// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

// +build integration

package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestGetQueueLengthNonExistentHost(t *testing.T) {
	_, err := getQueueLength("amqp://non-existent-host//", "")
	if err == nil {
		t.Fatalf("Error expected")
	}
	if got, want := err.Error(), "dial tcp: lookup non-existent-host"; !strings.HasPrefix(got, want) {
		t.Errorf("Expected err='%s', got: '%s'", want, got)
	}
}

func TestGetQueueLengthNonExistentQueue(t *testing.T) {
	_, err := getQueueLength(amqpURI(), "non-existent-queue")
	if err == nil {
		t.Fatalf("Error expected")
	}
	if got, want := err.Error(), `Exception (404) Reason: "NOT_FOUND - no queue 'non-existent-queue' in vhost '/'"`; got != want {
		t.Errorf("Expected err='%s', got: '%s'", want, got)
	}
}

func TestGetQueueLengthNoMsgs(t *testing.T) {
	conn, err := amqp.Dial(amqpURI())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	tmpQ, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	msgs, err := getQueueLength(amqpURI(), tmpQ.Name)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := msgs, 0; got != want {
		t.Errorf("Expected %d, got: %d", want, got)
	}

	_, err = ch.QueueDelete(tmpQ.Name, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetQueueLength(t *testing.T) {
	conn, err := amqp.Dial(amqpURI())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	tmpQ, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = ch.Publish(
			"",        // exchange
			tmpQ.Name, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				Body:          []byte(strconv.Itoa(i)),
				CorrelationId: strconv.Itoa(i),
				DeliveryMode:  amqp.Persistent,
			})
		if err != nil {
			t.Fatal(err)
		}
	}

	msgs, err := getQueueLength(amqpURI(), tmpQ.Name)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := msgs, 10; got != want {
		t.Errorf("Expected %d, got: %d", want, got)
	}

	_, err = ch.QueueDelete(tmpQ.Name, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMonitorQueueClosedChannel(t *testing.T) {
	forever := make(chan struct{}, 1)
	close(forever)

	f := func(i int) error {
		t.Fatalf("Unexpected result %d", i)
		return nil
	}

	monitorQueue("", []string{""}, 1, f, forever)
}

func TestMonitorQueue(t *testing.T) {
	conn, err := amqp.Dial(amqpURI())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	tmpQ, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = ch.Publish(
			"",        // exchange
			tmpQ.Name, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				Body:          []byte(strconv.Itoa(i)),
				CorrelationId: strconv.Itoa(i),
				DeliveryMode:  amqp.Persistent,
			})
		if err != nil {
			t.Fatal(err)
		}
	}

	forever := make(chan struct{}, 1)

	f := func(i int) error {
		if got, want := i, 10; got != want {
			t.Errorf("Expected %d, got: %d", want, got)
		}
		close(forever)
		return nil
	}

	monitorQueue(amqpURI(), []string{tmpQ.Name}, 1, f, forever)

	_, err = ch.QueueDelete(tmpQ.Name, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMonitorQueueWithTwoQueues(t *testing.T) {
	conn, err := amqp.Dial(amqpURI())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	tmpQ1, err := ch.QueueDeclare(
		"queue1", // name
		false,    // durable
		true,     // delete when usused
		false,    // exclusive
		false,    // noWait
		nil,      // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	tmpQ2, err := ch.QueueDeclare(
		"queue2", // name
		false,    // durable
		true,     // delete when usused
		false,    // exclusive
		false,    // noWait
		nil,      // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	queueNames := []string{tmpQ1.Name, tmpQ2.Name}

	for i := 0; i < 10; i++ {
		for _, name := range queueNames {
			err = ch.Publish(
				"",    // exchange
				name,  // routing key
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					Body:          []byte(strconv.Itoa(i)),
					CorrelationId: strconv.Itoa(i),
					DeliveryMode:  amqp.Persistent,
				})
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	forever := make(chan struct{}, 1)

	f := func(i int) error {
		if got, want := i, 20; got != want {
			t.Errorf("Expected %d, got: %d", want, got)
		}
		close(forever)
		return nil
	}

	monitorQueue(amqpURI(), queueNames, 1, f, forever)

	for _, name := range queueNames {
		_, err = ch.QueueDelete(name, false, false, true)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMonitorQueueGetQueueLengthError(t *testing.T) {

	forever := make(chan struct{}, 1)

	f := func(i int) error {
		t.Fatalf("Unexpected result %d", i)
		return nil
	}

	go monitorQueue("amqp://non-existent-host//", []string{"no-queue"}, 1, f, forever)

	time.Sleep(3 * time.Second)
	close(forever)

}

func TestMonitorQueueSaveStatError(t *testing.T) {
	conn, err := amqp.Dial(amqpURI())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	tmpQ, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		t.Fatal(err)
	}

	forever := make(chan struct{}, 1)

	f := func(i int) error {
		if got, want := i, 0; got != want {
			t.Errorf("Expected %d, got: %d", want, got)
		}
		return errors.New("Dummy error")
	}

	go monitorQueue(amqpURI(), []string{tmpQ.Name}, 1, f, forever)

	time.Sleep(3 * time.Second)
	close(forever)

	_, err = ch.QueueDelete(tmpQ.Name, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func amqpURI() string { return os.Getenv("AMQP_URI") }
