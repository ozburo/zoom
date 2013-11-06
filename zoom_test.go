// Copyright 2013 Alex Browne.  All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package zoom

import (
	"github.com/garyburd/redigo/redis"
	"reflect"
	"testing"
)

func TestSave(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	// create and save a basic model
	m := &basicModel{Attr: "Test"}
	err := Save(m)
	if err != nil {
		t.Error(err)
	}

	conn := GetConn()
	defer conn.Close()
	checkBasicModelSaved(t, m, conn)

	// TODO: test that the model was saved
}

func TestVariadicSave(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	models, err := newBasicModels(3)
	if err != nil {
		t.Error(err)
	}

	if err := Save(Models(models)...); err != nil {
		t.Error(err)
	}

	conn := GetConn()
	defer conn.Close()

	for _, m := range models {
		checkBasicModelSaved(t, m, conn)
	}
}

func TestFindById(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	m := &basicModel{Attr: "test"}
	if err := Save(m); err != nil {
		t.Error(err)
	}

	// use FindById to get a copy of the person
	result, err := FindById("basicModel", m.Id)
	if err != nil {
		t.Error(err)
	}
	mCopy, ok := result.(*basicModel)
	if !ok {
		t.Errorf("Could not convert type %T to *basicModel", result)
	}

	// make sure the found model is the same as original
	if !reflect.DeepEqual(m, mCopy) {
		t.Errorf("Found model did not match.\nExpected: %+v\nGot: %+v\n", m, mCopy)
	}
}

func TestScanById(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	m := &basicModel{Attr: "test"}
	if err := Save(m); err != nil {
		t.Error(err)
	}

	// use ScanById to get a copy of the person
	mCopy := new(basicModel)
	if err := ScanById(m.Id, mCopy); err != nil {
		t.Error(err)
	}

	// make sure the found model is the same as original
	if !reflect.DeepEqual(m, mCopy) {
		t.Errorf("Found model did not match.\nExpected: %+v\nGot: %+v\n", m, mCopy)
	}
}

func TestDelete(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	conn := GetConn()
	defer conn.Close()

	m := &basicModel{Attr: "test"}
	if err := Save(m); err != nil {
		t.Error(err)
	}
	checkBasicModelSaved(t, m, conn)

	if err := Delete(m); err != nil {
		t.Error(err)
	}

	// make sure it's gone
	checkBasicModelDeleted(t, m.Id, conn)
}

func TestVariadicDelete(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	models, err := newBasicModels(3)
	if err != nil {
		t.Error(err)
	}

	if err := Save(Models(models)...); err != nil {
		t.Error(err)
	}

	conn := GetConn()
	defer conn.Close()

	for _, m := range models {
		checkBasicModelSaved(t, m, conn)
	}

	if err := Delete(Models(models)...); err != nil {
		t.Error(err)
	}

	for _, m := range models {
		checkBasicModelDeleted(t, m.Id, conn)
	}
}

func TestDeleteById(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	conn := GetConn()
	defer conn.Close()

	m := &basicModel{Attr: "test"}
	if err := Save(m); err != nil {
		t.Error(err)
	}
	checkBasicModelSaved(t, m, conn)

	if err := DeleteById("basicModel", m.Id); err != nil {
		t.Error(err)
	}

	checkBasicModelDeleted(t, m.Id, conn)
}

func TestVariadicDeleteById(t *testing.T) {
	testingSetUp()
	defer testingTearDown()

	models, err := newBasicModels(3)
	if err != nil {
		t.Error(err)
	}

	if err := Save(Models(models)...); err != nil {
		t.Error(err)
	}

	conn := GetConn()
	defer conn.Close()

	for _, m := range models {
		checkBasicModelSaved(t, m, conn)
	}

	ids := make([]string, len(models))
	for i, m := range models {
		ids[i] = m.Id
	}
	if err := DeleteById("basicModel", ids...); err != nil {
		t.Error(err)
	}

	for _, m := range models {
		checkBasicModelDeleted(t, m.Id, conn)
	}
}

func checkBasicModelSaved(t *testing.T, m *basicModel, conn redis.Conn) {
	// make sure it was assigned an id
	if m.Id == "" {
		t.Error("model was not assigned an id")
	}

	// make sure the key exists
	key := "basicModel:" + m.Id
	exists, err := KeyExists(key, conn)
	if err != nil {
		t.Error(err)
	}
	if exists == false {
		t.Error("model was not saved in redis")
	}

	// make sure the attributes are correct for the model
	attr, err := redis.String(conn.Do("HGET", key, "Attr"))
	if err != nil {
		t.Error(err)
	}
	if attr != m.Attr {
		t.Errorf("Attr of saved model was incorrect. Expected: %s. Got: %s.\n", m.Attr, attr)
	}

	// make sure it was added to the basicModel:all index
	indexed, err := SetContains("basicModel:all", m.Id, conn)
	if err != nil {
		t.Error(err)
	}
	if indexed == false {
		t.Error("model was not added to basicModel:all")
	}
}

func checkBasicModelDeleted(t *testing.T, id string, conn redis.Conn) {
	key := "basicModel:" + id
	exists, err := KeyExists(key, conn)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Error("model key still exists")
	}

	// Make sure it was removed from index
	indexed, err := SetContains("basicModel:all", id, conn)
	if err != nil {
		t.Error(err)
	}
	if indexed {
		t.Error("model id is still in basicModel:all")
	}
}