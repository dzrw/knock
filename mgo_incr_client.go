package main

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"strconv"
	"time"
)

const (
	DEFAULT_MONGO_DATABASE      string = "bam"
	DEFAULT_MONGO_COLLECTION    string = "userdata"
	DEFAULT_MONGO_WRITE_CONCERN int    = 1
)

type mgo_incr_client struct {
	s              *mgo.Session
	properties     map[string]string
	url            string
	db             string
	collectionName string
	writeConcern   int
	deadbeef_id    interface{}
	fieldcount     int
}

type M bson.M

func (this *mgo_incr_client) Init(props map[string]string) (err error) {
	rand.Seed(time.Now().UnixNano())

	err = this.parseProperties(props)
	if err != nil {
		return
	}

	err = this.dial()
	if err != nil {
		return
	}

	err = this.plant_deadbeef_document()
	if err != nil {
		return
	}

	return
}

func (this *mgo_incr_client) Close() {
	this.s.Close()
}

func (this *mgo_incr_client) Work() (res WorkResult) {
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

func (this *mgo_incr_client) parseProperties(props map[string]string) (err error) {
	if v, ok := props["mongodb.url"]; ok {
		this.url = v
	} else {
		return errors.New("mongodb.url is a required property")
	}

	if v, ok := props["mongodb.database"]; ok {
		this.db = v
	} else {
		this.db = DEFAULT_MONGO_DATABASE
	}

	if v, ok := props["mongodb.writeConcern"]; ok {
		w, err := strconv.Atoi(v)
		if err != nil || w < 0 {
			return errors.New("mongodb.writeConcern must be >= 0")
		}

		this.writeConcern = w
	} else {
		this.writeConcern = DEFAULT_MONGO_WRITE_CONCERN
	}

	if v, ok := props["fieldcount"]; ok {
		u, err := strconv.Atoi(v)
		if err != nil || u <= 0 {
			return errors.New("fieldcount must be > 0")
		}

		this.fieldcount = u
	} else {
		this.fieldcount = 10
	}

	this.collectionName = DEFAULT_MONGO_COLLECTION
	return
}

func (this *mgo_incr_client) dial() (err error) {
	session, err := mgo.Dial(this.url)
	if err != nil {
		return
	}

	this.s = session
	this.s.SetSafe(&mgo.Safe{W: this.writeConcern})
	return
}

func (this *mgo_incr_client) plant_deadbeef_document() (err error) {
	doc := M{"$set": M{"stream_id": "deadbeef", "account_id": "test_1"}}

	coll := this.s.DB(this.db).C(this.collectionName)
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

func (this *mgo_incr_client) fieldnum() (i int) {
	return rand.Intn(this.fieldcount)
}

func (this *mgo_incr_client) randomfieldname() (name string) {
	return fmt.Sprintf("field-%d", this.fieldnum())
}

func (this *mgo_incr_client) increment_and_set_deadbeef() (updated int, err error) {
	doc := M{"$inc": M{"total": 1}, "$set": M{"account_id": "test_1"}}
	doc["$inc"].(M)[this.randomfieldname()] = 1

	coll := this.s.DB(this.db).C(this.collectionName)
	info, err := coll.Upsert(M{"stream_id": "deadbeef"}, doc)
	if err != nil {
		return
	}

	updated = info.Updated
	return
}
