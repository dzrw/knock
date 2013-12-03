knock
=====

A tool for measuring the throughput and response time of an arbitrary function (typically, a network request).  A *benchmark* library written in Go.

### Motivation

If you look around the Web, many benchmarks focus on the able to measure a variety of metrics from a single function (e.g. an HTTP server).

`knock` is interested in obtaining two metrics - throughput and response time - from an arbitrary function.

### Usage

```Bash
Usage:
  knock [OPTIONS]

Help Options:
  -h, --help=               Show this help message

Application Options:
  -c, --clients=CLIENTS     the number of individual load elements (0)
  -d, --duration=SECONDS    the number of seconds to run this benchmark (0)
  -v, --verbose
  -p=                       additional properties ({})
      --version             display version information (false)
      --client-stats        whether or not to track individual client statistics (false)
  -r, --runtime-profile=STR Go runtime profiles (e.g. cpu, memory, block, threadcount, or behavior-specifc) ({})
```

### Examples

```Bash
export KNOCK_URL=-p mongodb.url:mongodb://localhost:27017
export KNOCK_EXP_CONF=-p mongodb.run:writes -p mongodb.writeConcern:w=1 -p mongodb.doc_length:512
export KNOCK_REPORT_FILE=results/writes-512b-c4-w1.tsv

knock -c4 -d15 -v $KNOCK_URL $KNOCK_EXP_CONF > $KNOCK_REPORT_FILE
```

### Scripts

Although you can run knock directly from the command-line, it's currently easier to schedule an entire test plan from a script.  See the Ruby files in the /scripts folder for examples.


### FAQ

#### Why did you build this?

I wasn't able to find a MongoDB benchmarking tool that was able to examine the behavior of atomic increments under varying Write Concerns, so I wrote the tool myself in a way that is constrained enough (i.e. only 2 metrics) to be plausibly extensible.

### Similar Tools

* [JMeter](http://jmeter.apache.org/)
* [wrk](https://github.com/wg/wrk)
* [mongo-perf](https://github.com/mongodb/mongo-perf)
* [memtier_benchmark](https://github.com/GarantiaData/memtier_benchmark)
