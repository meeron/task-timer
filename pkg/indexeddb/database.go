package indexeddb

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type idbDatabase struct {
	value app.Value
}

func (db *idbDatabase) CreateObjectStore(name string, keyPath string) IDBObjectStore {
	store := db.value.Call("createObjectStore", name, map[string]any{"keyPath": keyPath})
	return &idbObjectStore{value: store}
}

func (db *idbDatabase) WriteTransaction(storeName string) IDBObjectStore {
	store := db.value.Call("transaction", storeName, "readwrite").Call("objectStore", storeName)
	return &idbObjectStore{value: store}
}

func (db *idbDatabase) ReadTransaction(storeName string) IDBObjectStore {
	store := db.value.Call("transaction", storeName, "readonly").Call("objectStore", storeName)
	return &idbObjectStore{value: store}
}
