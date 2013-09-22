package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
	updated, err := this.increment_and_set_deadbeef()
	switch {
	case err != nil:
		return WRK_ERROR
	case updated != 1:
		return WRK_WTF
	default:
		return WRK_OK
	}
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

func (this *mongodb_counters) increment_and_set_deadbeef() (updated int, err error) {
	doc := M{"$inc": M{"total": 1}, "$set": M{"account_id": "test_1"}}
	doc["$inc"].(M)[this.randomFieldName()] = 1

	info, err := this.collection().Upsert(M{"stream_id": "deadbeef"}, doc)
	if err != nil {
		return
	}

	updated = info.Updated
	return
}

func (this *mongodb_counters) randomFieldName() (name string) {
	return randomFieldName(this.conf.fieldcount)
}
