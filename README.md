knock
=====

A tool for measuring the throughput and response time of an arbitrary function (typically, a network request).  A *benchmark* library written in Go.

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

### Why?

If you look around the Web, it seems like most benchmarks are able to measure a staggering number of metrics from a specific function (e.g. an HTTP server).

`knock` is interested in obtaining two metrics from any conceivable function.

### No, really, what are you up to?

I couldn't find a decent MongoDB benchmarking tool, and certainly not one that examined the behavior of atomic increments under varying Write Concerns, so I wrote the tool myself in a way that is sufficiently stupid enough to be plausibly extensible.

### Why didn't you use JMeter?

`knock` is 4MB, *without any other dependencies*, a program that I can just drop on a box and use the moment after the bits land.  The Go runtime is conceptually simple enough to reason about that doing a performance analysis of knock itself isn't a long-term team effort.  Understanding the performance characteristics of the tool itself is a big part of getting good numbers.

I don't trust Java with its 500GB of downloads, its myriad of GC-tuning command-line options, and its JIT compiler. With a 60s startup time. There, I said it.

### But, I like Java.

Oh, then, you'll probably find what you're looking for [here](http://jmeter.apache.org/).

