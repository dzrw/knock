// GOMAXPROCS == os.MaxCores

// TIME=30 minutes (ACTUAL=1 HOUR)
// CLI variables
//
// -c, --clients CLIENTS, default=1
// -g, --goroutines GOROUTINES, default=1
// -d, --duration TIMEOUT_SECONDS, duration of test, default=notset
// -v, --verbose (boolean) whether or not to write summary information to os.Stderr
// -p, --property key:value (passed as a map[string]string to behaviors)

// TIME=1 HOUR (ACTUAL=3 HOURS)
// a behavior:
// control load by varying the number of distinct sessions (each session has its own conn pool)
// hard-coded to use mgo with the counter increment bm for now
// potentially hard-coded to use a specific server
//    - insert a single document at init
/*

type M map[string]interface{}
doc := M{"_id": bson.NewObjectId(), "stream_id": "deadbeef"}
c := database.C("ycsb").Insert(doc)
*/

//    - for each operation:
//        - pick a random number 0-n
/*
      @db.collection("microbenchmark").update(
            {:test_id => test_id},
            {"$inc" => {
              "count.c#{n}.a" => 1,
              "count.total" => 1
            },
             '$set' => {"account_id" => account}},
            :upsert => true
          )

mgo:
>> db.links.update({"link":"www.example.com"}, {$set:{"slink":"g.g/1"},
>> $inc:{"count": 1}}, {upsert:true} )

type M map[string]interface{}

selector := M{"stream_id": "deadbeef"},
diff := M{"$set": M{"account_id": "test_1"},
          "$inc": M{fmt.Sprintf("count-%d", n): 1,
                    "total": 1}
          }

c := database.C("ycsb").Upsert(selector, diff)


*/
//    - delete the document at cleanup
/*
  id = this.doc_id
  database.C("ycsb").RemoveId(id)
*/
//
//
// Init(props map[string]string) Initialize any state for this DB.
// Cleanup() Cleanup any state for this DB.
// Operate() Perform a random unit of work (1 op) against the DB
//   (a Read, a Write, etc)

// TIME=1 HOUR (ACTUAL=1.5 HOURS)
// slaves:
// slaves perform operations until the test duration expires, then they terminate
// stream <slaveid,latency_msec> back to master over a fixed number of W channels (W=1)
// partition sessions into individual goroutines ("slaves")
//  - has a timeout/duration
//  - has more than 1 client
//  - is-a goroutine
//  - performs the client protocol (init, work, close)
//  - decrements a waitgroup when the duration/timeout elapses

// TIME=1 HOUR (stats funnel, waitgroup) (ACTUAL= ~1 HOUR)
// observer:
// spawn a goroutine with a tomb and a waitgroup to monitor slaves
// observer exposes a Receive() <-chan for master to listen for stats, and
//  a Report(slaveid, latency) method for slaves to write stats.
//  internally, this wraps around an unbuffered channel that can later be
//  expanded into a bufferred or elastic channel (or multiple channels)
//  if it becomes a bottleneck.

// TIME=1 HOUR (spawn slaves/observer, compute stats (avgs, ops/sec, hist))
// master:
// master distributes CLIENTS over slaves ("gorotines"), hands slaves a waitgroup
// master waits on the observer to know when to proceed to reporting
// master streams <average latency(msec), ops/sec> every 2 seconds to caller
// master builds a histogram of latencies <msec, count>
// master terminates out after the run completes
//

// TIME=1 HOUR (parse args, for/select, print report)
// caller:
// caller parses CLI args and starts master
// caller writes streaming summaries from master to os.Stderr (using ^M trick)
// caller waits for master to die (tomb)
// caller can read the final histogram, avg latency, and ops/sec from (dead) master
// caller writes final values to os.Stdout (write can be redirected to a file from the CLI).

/*

Every 2 seconds,

  master computes the new weighted average latency(msec) and the new ops/sec

  start_time = time.Now()

  at t=0, last_avg_lag = 0ms,  last_ops = 0  ==> 0ms lag, 0 ops/sec

  ...stream results 0<t<2 => 20 ops, sum_latency=1000ms

  at t=2,
    - compute the intermediate avg lag = 1000ms / 20 ops = 50ms/op
    - compute the weights of the old and new contributions:
        -- 0 ops / 0 ops + 20 ops == 0.0
        -- 20 ops / 0 ops + 20 ops == 1.0
        -- 0.0 + 1.0 == 1.0 (check)

    - computed the weighted average
        -- (0.0 * 0ms) + (1.0 * 50ms) == 50ms

    - compute the ops/sec:
        -- (0 ops + 20 ops) / (time.Since(start_time) = 20 ops / 2s == 10 ops/sec

    - update the counters:
        -- last_avg_lag = 50ms, last_ops = (0 + 20) = 20 ops

  ...stream results 0<t<2 => 100 ops, sum_latencies=1500ms

  at t=4,
    - stream avg lag == 1500ms / 100 ops = 15ms/op

    - weights:
      a) (20 ops / (20 ops + 100 ops)) == 0.1666
      b) (100 ops / (20 ops + 100 ops)) == 0.8333

    - weighted average latency:
      a) (0.1666 * 50ms) + (0.8333 * 15ms) == 8.3333ms + 12.5ms == 20.83ms

    - ops/sec:
      a) (20 ops + 100 ops) / (time.Since(start_time)) == 120 ops / 4s = 30 ops/sec

    - update counters:
      a) last_avg_lag = 20.83ms
      b) last_ops = (20 ops + 100 ops) == 120 ops

  ...

*/

/*
  - fix the clients over goroutines bug (ACTUAL= 10 min)

  - switch to 1 client per goroutine (ESTIMATED=30 min, ACTUAL=40min)
    -- the multiplexing is unnecessary and confusing
    -- the efficiency drops anyway
    -- bad assumption that it would be faster/better than the go scheduler

  - finalize the generated reports
    -- move stats tracking out of master (ACTUAL=1 HOUR)

    -- runtime/latency/throughput/planned load/active load (ACTUAL=30 min)
      -- This is summary stats written to stdout at the end of the round.
      -- might want to include conf details
      -- this is like the header of the report

    -- latency histogram (ACTUAL=30 min)
      -- One column per client id if tracking per client stats
      -- First column is overall

    -- throughput vs latency (XR curve) (PUNTED)
          -- Nope. This only makes sense if you've got multiple rounds.
             We've only got 1 round currently.

  - finalize summary output (ACTUAL=20 MIN)
    -- fix bug where the last second/final numbers aren't displayed (FIXED=5 min)

  - add properties for tuning internal channel buffer sizes (ACTUAL=30 min)
    -- specifically, the taskmaster's LatencyEventChannel.
*/

/*
  TODO
    - report the min and max response times
    - also the 5th, 95th, and 99th percentiles
*/
package main
