package main

import (
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
)

const (
	MONGO_MIN_DOCUMENT_LENGTH     = 64
	MONGO_DEFAULT_DOCUMENT_LENGTH = MONGO_MIN_DOCUMENT_LENGTH
)

type mongodb_writes struct {
	conf        *MongoBehaviorInfo
	deadbeef_id interface{}
	collection  func() *mgo.Collection
	doc_length  int
	doc_data    string
}

func (this *mongodb_writes) Init(info *MongoBehaviorInfo) (err error) {
	this.conf = info
	this.collection = info.collection

	err = this.parseProperties(this.conf.properties)
	if err != nil {
		return
	}

	this.doc_data = this.document_data()
	return
}

func (this *mongodb_writes) Close() {
	// nop
}

func (this *mongodb_writes) Work() (res WorkResult) {
	err := this.insert_document()
	switch {
	case err != nil:
		return WRK_ERROR
	default:
		return WRK_OK
	}
}

func (this *mongodb_writes) insert_document() (err error) {
	// Use the same document data every time to eliminate the
	// overhead of random data generation from the results.
	// Hopefully, that doesn't invalidate the test.

	doc := M{"_id": bson.NewObjectId(), "data": this.doc_data}
	err = this.collection().Insert(doc)
	if err != nil {
		return
	}

	return
}

func (this *mongodb_writes) document_data() string {
	const (
		fudge = 42
	)

	if this.doc_length == MONGO_MIN_DOCUMENT_LENGTH {
		return "I wrote some Go!" // should be exactly 64 bytes total now.
	} else {
		return strings.Repeat("x", this.doc_length-fudge)
	}
}

func (this *mongodb_writes) parseProperties(props map[string]string) (err error) {
	if v, ok := props["mongodb.doc_length"]; ok {
		w, err := strconv.Atoi(v)
		if err != nil || w < MONGO_MIN_DOCUMENT_LENGTH {
			return errors.New("mongodb.doc_length must be >= 64 bytes")
		}

		this.doc_length = w
	} else {
		this.doc_length = MONGO_DEFAULT_DOCUMENT_LENGTH
	}

	return
}
