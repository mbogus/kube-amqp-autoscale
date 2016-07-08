// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type queueMetrics struct {
	Count    int
	Average  float64
	Coverage float64
}

func dbPath(dir string, file string) (string, error) {
	if strings.ToLower(file) == ":memory:" {
		return file, nil
	}
	if dir == "" || file == "" {
		return "", errors.New("Missing directory and/or filename for the metrics database")
	}
	fi, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("Valid directory name is required, got '%s'", dir)
	}
	filename := filepath.Join(dir, file)
	if !isValidFile(filename) {
		return "", fmt.Errorf("Invalid database filename '%s'", filename)
	}
	return filename, nil
}

func isValidFile(path string) bool {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return true
	}

	// Attempt to create it
	var d []byte
	if err := ioutil.WriteFile(path, d, 0644); err == nil {
		os.Remove(path) // And delete it
		return true
	}

	return false
}

func connectToDB(path *string) (*sql.DB, error) {
	return sql.Open("sqlite3", *path)
}

func createTable(db *sql.DB) error {
	stmt, err := db.Prepare(createTableSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	return err
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS timeline (
	unix_secs INTEGER(4) PRIMARY KEY DESC NOT NULL DEFAULT (strftime('%s', 'now')),
	q_len INTEGER NOT NULL DEFAULT (0)
)`

const statsQuerySQL = `SELECT
	COALESCE(COUNT(1), 0) cnt, COALESCE(AVG(q_len), 0.0) average
FROM
	timeline
WHERE
	strftime('%s', 'now') - unix_secs <= ?`

const savePointSQL = `INSERT INTO timeline (q_len) VALUES (?)`
const deleteMetricsSQL = `DELETE FROM timeline WHERE strftime('%s', 'now') - unix_secs > ?`

func updateMetrics(db *sql.DB, count, duration int) error {
	if err := deleteMetrics(db, duration); err != nil {
		return err
	}
	if err := saveMetric(db, count); err != nil {
		return err
	}
	return nil
}

func deleteMetrics(db *sql.DB, duration int) error {
	stmt, err := db.Prepare(deleteMetricsSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(duration)
	return err
}

func saveMetric(db *sql.DB, count int) error {
	stmt, err := db.Prepare(savePointSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(count)
	return err
}

// getMetrics returns number of metrics, average queue length over specified
// period of time (in seconds)
func getMetrics(db *sql.DB, duration, interval int) (*queueMetrics, error) {
	stmt, err := db.Prepare(statsQuerySQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(duration)

	metrics := queueMetrics{}
	row.Scan(&metrics.Count, &metrics.Average)
	metrics.Coverage = float64(metrics.Count) * float64(interval) / float64(duration)
	return &metrics, nil
}
