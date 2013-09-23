package main

import (
	"labix.org/v2/mgo/bson"
	_ "log"
	_ "strconv"
	"testing"
	"time"
)

func TestClientInit(t *testing.T) {
	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run": "counters",
		"mongodb.url": "mongodb://localhost:27017",
	}

	err := client.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	defer client.Close()

	docid := client.mb.(*mongodb_counters).deadbeef_id
	coll := client.s.DB(client.db).C(client.collectionName)
	query := coll.FindId(docid)

	n, err := query.Count()
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, 1, n) {
		return
	}
}

func TestClientWork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run": "counters",
		"mongodb.url": "mongodb://localhost:27017",
	}

	// Connect to mongo
	err := client.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	// Remember to close
	defer client.Close()

	// Find the baseline total number
	var res struct {
		Id        bson.ObjectId `bson:"_id"`
		Total     int64         `bson:"total"`
		AccountId string        `bson:"account_id"`
	}

	docid := client.mb.(*mongodb_counters).deadbeef_id
	coll := client.s.DB(client.db).C(client.collectionName)
	err = coll.FindId(docid).One(&res)
	if !expectOk(t, err) {
		return
	}

	orig_total := int(res.Total)

	// Do several units of work
	var wr WorkResult
	for i := 0; i < 20; i += 1 {
		wr = client.Work(time.Now())
		if wr != WRK_OK {
			t.Errorf("expected: WRK_OK, got: %v", wr)
			return
		}
	}

	// Check that the total has increased by the same amount
	err = coll.FindId(docid).One(&res)
	if !expectOk(t, err) {
		return
	}

	next_total := int(res.Total)

	if !expectInt(t, orig_total+20, next_total) {
		return
	}
}
