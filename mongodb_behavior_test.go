package main

import (
	_ "labix.org/v2/mgo/bson"
	_ "log"
	_ "strconv"
	"testing"
	_ "time"
)

func TestMongoDbBehaviorProperties(t *testing.T) {
	client := &mongodb_behavior{}
	props := make(map[string]string)

	err := client.parseProperties(props)
	if err == nil {
		t.Error("empty properties should raise an error")
		return
	}

	props["mongodb.run"] = "counters"
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

func TestMongoDbDialInternals(t *testing.T) {
	client := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run":          "counters",
		"mongodb.url":          "mongodb://localhost:27017",
		"mongodb.writeConcern": "0",
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
