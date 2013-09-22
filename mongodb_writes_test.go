package main

import (
	_ "labix.org/v2/mgo/bson"
	_ "log"
	"strconv"
	"testing"
	_ "time"
)

func TestMongoDbWritesProperties(t *testing.T) {
	mode := &mongodb_writes{}
	props := make(map[string]string)

	err := mode.parseProperties(props)
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, MONGO_DEFAULT_DOCUMENT_LENGTH, mode.doc_length) {
		return
	}

	props["mongodb.doc_length"] = "NAN"

	err = mode.parseProperties(props)
	if err == nil {
		t.Error("expected an error when mongodb.doc_length is not a number")
		return
	}

	props["mongodb.doc_length"] = strconv.Itoa(MONGO_DEFAULT_DOCUMENT_LENGTH - 1)

	err = mode.parseProperties(props)
	if err == nil {
		t.Error("expected an error when mongodb.doc_length < 64")
		return
	}

	props["mongodb.doc_length"] = "64"

	err = mode.parseProperties(props)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, MONGO_DEFAULT_DOCUMENT_LENGTH, mode.doc_length) {
		return
	}
}

func TestMongoDbWritesInit(t *testing.T) {
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

	mode, ok := b.mb.(*mongodb_writes)
	if !expectBool(t, true, ok) {
		return
	}

	const content = "I wrote some Go!"
	if !expectString(t, content, mode.doc_data) {
		return
	}
}

func TestMongoDbWritesWork(t *testing.T) {
	b := &mongodb_behavior{}
	props := map[string]string{
		"mongodb.run": "writes",
		"mongodb.url": "mongodb://localhost:27017",
	}

	// Connect to mongo
	err := b.Init(props)
	if err != nil {
		t.Error(err)
		return
	}

	// Remember to close
	defer b.Close()

	mode, ok := b.mb.(*mongodb_writes)
	if !expectBool(t, true, ok) {
		return
	}

	// Find the baseline total number of documents in this collection
	baseline, err := mode.collection().Count()
	if !expectOk(t, err) {
		return
	}

	// Do several units of work
	for i := 0; i < 20; i += 1 {
		wr := b.Work()
		if wr != WRK_OK {
			t.Errorf("expected: WRK_OK, got: %v", wr)
			return
		}
	}

	// Check that the number of documents has increased by the same amount
	v, err := mode.collection().Count()
	if !expectOk(t, err) {
		return
	}

	if !expectInt(t, baseline+20, v) {
		return
	}
}
