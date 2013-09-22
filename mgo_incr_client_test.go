package main

import (
	"labix.org/v2/mgo/bson"
	_ "log"
	_ "strconv"
	"testing"
	_ "time"
)

func TestClientProperties(t *testing.T) {
	client := &mongodb_behavior{}
	props := make(map[string]string)

	err := client.parseProperties(props)
	if err == nil {
		t.Error("empty properties should raise an error")
		return
	}

	props["mongodb.behavior"] = "counters"
	props["mongodb.url"] = "bad_url_for_testing"

	err = client.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectString(t, DEFAULT_MONGO_DATABASE, client.db) {
		return
	}

	if !expectInt(t, DEFAULT_MONGO_WRITE_CONCERN, client.writeConcern) {
		return
	}

	if !expectInt(t, 10, client.fieldcount) {
		return
	}

	if !expectString(t, DEFAULT_MONGO_COLLECTION, client.collectionName) {
		return
	}

	props["mongodb.database"] = "bad_db"

	err = client.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectString(t, "bad_db", client.db) {
		return
	}

	props["mongodb.writeConcern"] = "NAN"

	err = client.parseProperties(props)
	if err == nil {
		t.Error("expected an error when mongodb.writeConcern is not a number")
		return
	}

	props["mongodb.writeConcern"] = "-1"

	err = client.parseProperties(props)
	if err == nil {
		t.Error("expected an error when mongodb.writeConcern < 0")
		return
	}

	props["mongodb.writeConcern"] = "42"

	err = client.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 42, client.writeConcern) {
		return
	}

	props["fieldcount"] = "NAN"

	err = client.parseProperties(props)
	if err == nil {
		t.Error("expected an error when fieldcount is not a number")
		return
	}

	props["fieldcount"] = "0"

	err = client.parseProperties(props)
	if err == nil {
		t.Error("expected an error when fieldcount <= 0")
		return
	}

	props["fieldcount"] = "42"

	err = client.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 42, client.fieldcount) {
		return
	}
}

func TestClientDialInternals(t *testing.T) {
	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.behavior": "counters",
		"mongodb.url":      "mongodb://localhost:27017",
	}

	err := client.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	err = client.dial()
	if err != nil {
		t.Error(err)
		return
	}

	defer client.s.Close()

	if !expectInt(t, client.writeConcern, client.s.Safe().W) {
		return
	}
}

func TestClientInit(t *testing.T) {
	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.behavior": "counters",
		"mongodb.url":      "mongodb://localhost:27017",
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
	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.behavior": "counters",
		"mongodb.url":      "mongodb://localhost:27017",
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
		wr = client.Work()
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
