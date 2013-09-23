package main

const (
	/*
	 CHANGES

	 1.1.1:
	   * added "internals.OpsPerStall" property to insert a 1s
	     pause into every sandbox after this many operations.
	   * fixed some more bugs related to measuring very small
	     latencies

	 1.1.0:
	   * added a second MongoDB experiment "writes" for writing
	     random-ish documents of a fixed length to the database

	 1.0.0:
	   * initial drop containing the "counters" experiment for
	     MongoDB which performs increments of counters on a
	     single document
	*/
	VERSION = "1.1.1"
)
