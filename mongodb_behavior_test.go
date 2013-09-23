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

	props["mongodb.writeConcern"] = "foo"

	err = client.parseProperties(props)
	if err == nil {
		t.Error("expected an error when mongodb.writeConcern is not supported")
		return
	}

	props["mongodb.writeConcern"] = "none"

	err = client.parseProperties(props)
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, -1, client.writeConcern) {
		return
	}

	props["mongodb.writeConcern"] = "w=0"

	err = client.parseProperties(props)
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, 0, client.writeConcern) {
		return
	}

	props["mongodb.writeConcern"] = "w=1"

	err = client.parseProperties(props)
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, 1, client.writeConcern) {
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
		"mongodb.run": "counters",
		"mongodb.url": "mongodb://localhost:27017",
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
}

func TestMongoDbWriteConcernDefault(t *testing.T) {
	b := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run": "writes",
		"mongodb.url": "mongodb://localhost:27017",
	}

	err := b.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	defer b.Close()

	if !expectInt(t, DEFAULT_MONGO_WRITE_CONCERN, b.s.Safe().W) {
		return
	}
}

func TestMongoDbWriteConcern0(t *testing.T) {
	b := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run":          "writes",
		"mongodb.url":          "mongodb://localhost:27017",
		"mongodb.writeConcern": "w=0",
	}

	err := b.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	defer b.Close()

	if !expectInt(t, 0, b.s.Safe().W) {
		return
	}
}

func TestMongoDbWriteConcern1(t *testing.T) {
	b := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run":          "writes",
		"mongodb.url":          "mongodb://localhost:27017",
		"mongodb.writeConcern": "w=1",
	}

	err := b.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	defer b.Close()

	if !expectInt(t, 1, b.s.Safe().W) {
		return
	}
}

func TestMongoDbWriteConcernUnacknowledged(t *testing.T) {
	b := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run":          "writes",
		"mongodb.url":          "mongodb://localhost:27017",
		"mongodb.writeConcern": "none",
	}

	err := b.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	defer b.Close()

	if b.s.Safe() != nil {
		t.Error("expected nil, got: %v", b.s.Safe())
		return
	}
}
