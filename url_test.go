package main

import (
	"github.com/streadway/amqp"
	"testing"
)

func TestUnquoteURI_localhost_ipv4(t *testing.T) {
	uri := "amqp://guest:guest@127.0.0.1:5672//"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_ipv6(t *testing.T) {
	uri := "amqp://guest:guest@[::1]:5672//"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost(t *testing.T) {
	uri := "amqp://guest:guest@localhost:5672//"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_quoted(t *testing.T) {
	uri := "'amqp://guest:guest@localhost:5672//'"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_ipv4_quoted(t *testing.T) {
	uri := "'amqp://guest:guest@127.0.0.1:5672//'"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_ipv6_quoted(t *testing.T) {
	uri := "'amqp://guest:guest@[::1]:5672//'"
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_doublequoted(t *testing.T) {
	uri := "\"amqp://guest:guest@localhost:5672//\""
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_ipv4_doublequoted(t *testing.T) {
	uri := "\"amqp://guest:guest@127.0.0.1:5672//\""
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}

func TestUnquoteURI_localhost_ipv6_doublequoted(t *testing.T) {
	uri := "\"amqp://guest:guest@[::1]:5672//\""
	_, err := amqp.ParseURI(unquoteURI(uri))
	if err != nil {
		t.Errorf(err.Error(), err)
	}
}