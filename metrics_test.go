// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestDbPathMemory(t *testing.T) {
	db, err := dbPath("", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := db, ":memory:"; got != want {
		t.Errorf("Expected dbPath='%s', got: '%s'", want, got)
	}
}

func TestDbPathNoFileDir(t *testing.T) {
	_, err := dbPath("", "")
	if err == nil {
		t.Fatal("Error expected")
	}
	if got, want := err.Error(), "missing directory and/or filename for the metrics database"; got != want {
		t.Errorf("Expected dbPath='%s', got: '%s'", want, got)
	}
}

func TestDbPathTempFile(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	db, err := dbPath(path.Dir(file.Name()), path.Base(file.Name()))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := db, file.Name(); got != want {
		t.Errorf("Expected dbPath='%s', got: '%s'", want, got)
	}
}

func TestDbPathNewFile(t *testing.T) {
	fname := strconv.Itoa(rand.Int())
	db, err := dbPath(os.TempDir(), fname)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := db, path.Join(os.TempDir(), fname); got != want {
		t.Errorf("Expected dbPath='%s', got: '%s'", want, got)
	}
}

func TestCreateTable(t *testing.T) {
	testDBFile := ":memory:"
	db, err := connectToDB(&testDBFile)
	if err != nil {
		t.Fatal(err)
	}
	err = createTable(db)
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare(`SELECT unix_secs, q_len FROM timeline`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if got, want := rows.Next(), false; got != want {
		t.Errorf("Expected has rows='%v', got: '%v'", want, got)
	}
}

func TestUpdateMetrics(t *testing.T) {
	testDBFile := ":memory:"
	db, err := connectToDB(&testDBFile)
	if err != nil {
		t.Fatal(err)
	}
	err = createTable(db)
	if err != nil {
		t.Fatal(err)
	}

	duration := 2
	for i := 0; i < 30; i++ {
		updateMetrics(db, i, duration)
		time.Sleep(100 * time.Millisecond)
	}
	stmt, err := db.Prepare(`SELECT MIN(unix_secs) FROM timeline`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var minTime int64
	row.Scan(&minTime)

	if got, want := minTime, time.Now().Unix(); got+int64(duration) > want {
		t.Errorf("Expected min date='%v', got: '%v'", want, got)
	}
}

func TestGetMetrics(t *testing.T) {
	testDBFile := ":memory:"
	db, err := connectToDB(&testDBFile)
	if err != nil {
		t.Fatal(err)
	}
	err = createTable(db)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		updateMetrics(db, i, 6)
		time.Sleep(1 * time.Second)
	}
	stats, err := getMetrics(db, 6, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := stats.Count, 3; got != want {
		t.Errorf("Expected count='%v', got: '%v'", want, got)
	}
	if got, want := stats.Average, 1.0; got != want {
		t.Errorf("Expected average='%v', got: '%v'", want, got)
	}
	if got, want := stats.Coverage, 0.5; got != want {
		t.Errorf("Expected coverage='%v', got: '%v'", want, got)
	}

}
