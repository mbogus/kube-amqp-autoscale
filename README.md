# kube-amqp-autoscale

Dynamically scale kubernetes resources using length of an AMQP queue (number of messages available for retrieval from the queue) to determine the load on an application/Kubernetes pod.


**NOTICE**

If your application load is not queue-bound but rather CPU-sensitive, make sure to use built-in Kubernetes [Horizontal Pod Autoscaling](http://kubernetes.io/docs/user-guide/horizontal-pod-autoscaling/) instead of this project.

## Status

*Alpha*

[![Build Status](https://travis-ci.org/mbogus/kube-amqp-autoscale.svg?branch=master)](https://travis-ci.org/mbogus/kube-amqp-autoscale)  [![Coverage Status](https://coveralls.io/repos/github/mbogus/kube-amqp-autoscale/badge.svg?branch=master)](https://coveralls.io/github/mbogus/kube-amqp-autoscale?branch=master)


## Usage

The best way to use the service is by importing [automated build docker container](https://hub.docker.com/r/mbogus/kube-amqp-autoscale/).
Please follow the instructions on the automated build page.


## Go get

    go get github.com/mbogus/kube-amqp-autoscale


## Clone from [github](https://github.com/mbogus/kube-amqp-autoscale)

* Create directory for APT projects `mkdir -p $GOPATH/src/github.com/mbogus`
  as typical in [writing go programs](https://golang.org/doc/code.html)
* Clone this project `git clone https://github.com/mbogus/kube-amqp-autoscale.git`


### Building on Windows
If you have a unix-y shell on Windows ([MSYS2](http://sourceforge.net/p/msys2/wiki/MSYS2%20installation/),
[CYGWIN](https://cygwin.com/install.html) or other), see *Build project* below.


### Build project

The project depends on several external Go projects that can be automatically
downloaded using `make depend` target.

Run `make depend && make [build]`


## Runtime environment variables

* `GOMAXPROCS` limits the number of operating system threads that can execute
  user-level Go code simultaneously


## Runtime command-line arguments

* **`amqp-uri`** required, RabbitMQ broker URI, e.g. `amqp://guest:guest@127.0.0.1:5672//`
* **`amqp-queue`** required, RabbitMQ queue name to measure load on an application.  Use comma separator to specify multiple queues.
* **`api-url`** required, Kubernetes API URL, e.g. `http://127.0.0.1:8080`
* `api-user` optional, username for basic authentication on Kubernetes API
* `api-passwd` optional, password for basic authentication on Kubernetes API
* `api-token` optional, path to a bearer token file for OAuth authentication, on a Kubernetes pod usually `/var/run/secrets/kubernetes.io/serviceaccount/token`
* `api-cafile` optional, path to CA certificate file for HTTPS connections to Kubernetes API from within a cluster, typically `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`
* `api-insecure` optional, set to `true` for connecting to Kubernetes API without verifying TLS certificate; unsafe, use for development only (default `false`)
* `min` lower limit for the number of replicas for a Kubernetes pod that can be set by the autoscaler (default `1`)
* **`max`** required, upper limit for the number of replicate for a Kubernetes pod that can be set by the autoscaler (must be greater than `min`)
* **`name`** required, name of the Kubernetes resource to autoscale
* `kind` type of the Kubernetes resource to autoscale, one of `Deployment`, `ReplicationController`, `ReplicaSet` (default `Deployment`)
* `ns` Kubernetes namespace (default `default`)
* `interval` time interval between Kubernetes resource scale runs in secs (default `30`)
* **`threshold`** required, number of messages on a queue representing maximum load on the autoscaled Kubernetes resource
* `increase-limit` limit number of Kubernetes pods to be provisioned in a single scale iteration to max of the value, set to a number greater than 0, default `unbounded`
* `decrease-limit` limit number of Kubernetes pods to be terminated in a single scale iteration to max of the value, set to a number greater than 0, default `unbounded`
* `stats-interval` time interval between metrics gathering runs in seconds (default `5`)
* `eval-intervals` number of autoscale intervals used to calculate average queue length (default `2`)
* `stats-coverage` required percentage of statistics to calculate average queue length (default `0.75`)
* `db` sqlite3 database filename for storing  queue length statistics (default `file::memory:?cache=shared`)
* `db-dir` directory for sqlite3 statistics database file
* `version` show version
* `metrics-listen-address` the address to listen on for exporting Prometheus metrics (default `:9505`)


## Integration tests

To run intergation tests, make sure to configure access to running RabbitMQ instance,
export environment variable `AMQP_URI=amqp://username:passwd@rabbitmq-host:5672//`
and run `go test -tags=integration ./...`
