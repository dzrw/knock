package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

type mongodb_counters struct {
	conf        *MongoBehaviorInfo
	deadbeef_id interface{}
	collection  func() *mgo.Collection
}

func (this *mongodb_counters) Init(info *MongoBehaviorInfo) (err error) {
	this.conf = info
	this.collection = info.collection

	err = this.plant_deadbeef_document()
	if err != nil {
		return
	}

	return
}

func (this *mongodb_counters) Close() {
	// nothing to do
}

func (this *mongodb_counters) Work() (res WorkResult) {
	doc := M{"$inc": M{"total": 1}, "$set": M{"account_id": "test_1"}}
	doc["$inc"].(M)[this.randomFieldName()] = 1

	info, err := this.collection().Upsert(M{"stream_id": "deadbeef"}, doc)

	switch {
	case err != nil:
		log.Fatalf("counters: %+v", err)
	case info != nil:
		res = WRK_OK
	case this.conf.writeConcern == -1:
		res = WRK_OK
	default:
		res = WRK_WTF
	}

	return
}

func (this *mongodb_counters) plant_deadbeef_document() (err error) {
	doc := M{"$set": M{"stream_id": "deadbeef", "account_id": "test_1"}}
	coll := this.collection()

	_, err = coll.Upsert(M{"stream_id": "deadbeef"}, doc)
	if err != nil {
		return
	}

	var res struct {
		Id bson.ObjectId `bson:"_id"`
	}

	err = coll.Find(M{"stream_id": "deadbeef"}).One(&res)
	if err != nil {
		return
	}

	this.deadbeef_id = res.Id
	return
}

func (this *mongodb_counters) randomFieldName() (name string) {
	return randomFieldName(this.conf.fieldcount)
}
