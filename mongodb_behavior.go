package main

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"strconv"
)

const (
	DEFAULT_MONGO_DATABASE      string = "knock"
	DEFAULT_MONGO_COLLECTION    string = "userdata"
	DEFAULT_MONGO_WRITE_CONCERN int    = 1
)

type MongoBehaviorInfo struct {
	session      *mgo.Session
	writeConcern int
	fieldcount   int
	properties   map[string]string
	collection   func() *mgo.Collection
}

type MongoBehavior interface {
	Init(info *MongoBehaviorInfo) (err error)
	Close()
	Work() (res WorkResult)
}

type mongodb_behavior struct {
	s              *mgo.Session
	properties     map[string]string
	url            string
	db             string
	collectionName string
	writeConcern   int
	fieldcount     int
	mb             MongoBehavior
}

type M bson.M

func (this *mongodb_behavior) Init(props map[string]string) (err error) {
	err = this.parseProperties(props)
	if err != nil {
		return
	}

	err = this.dial()
	if err != nil {
		return
	}

	info := &MongoBehaviorInfo{
		this.s,
		this.writeConcern,
		this.fieldcount,
		this.properties,
		func() *mgo.Collection {
			return this.s.DB(this.db).C(this.collectionName)
		},
	}

	err = this.mb.Init(info)
	if err != nil {
		return
	}

	return
}

func (this *mongodb_behavior) Close() {
	defer this.s.Close()
	this.mb.Close()
}

func (this *mongodb_behavior) Work() (res WorkResult) {
	return this.mb.Work()
}

func (this *mongodb_behavior) parseProperties(props map[string]string) (err error) {
	if v, ok := props["mongodb.behavior"]; ok {
		switch v {
		case "counters":
			this.mb = &mongodb_counters{}
		// case "writeconcern":
		// 	this.mb = &mongodb_writeconcern{}
		default:
			return errors.New("mongodb.behavior must be one of counters, writeconcern")
		}
	} else {
		return errors.New("mongodb.behavior is a required property")
	}

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

func (this *mongodb_behavior) dial() (err error) {
	session, err := mgo.Dial(this.url)
	if err != nil {
		return
	}

	this.s = session
	this.s.SetSafe(&mgo.Safe{W: this.writeConcern})
	return
}

func randomFieldName(fieldcount int) (name string) {
	return fmt.Sprintf("field-%d", rand.Intn(fieldcount))
}
