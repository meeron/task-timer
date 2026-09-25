package indexeddb

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type idbDatabase struct {
	value app.Value
}

func (db *idbDatabase) CreateObjectStore(name string, keyPath string) IDBObjectStore {
	store := db.value.Call("createObjectStore", name, map[string]interface{}{"keyPath": keyPath})
	return &idbObjectStore{value: store}
}
