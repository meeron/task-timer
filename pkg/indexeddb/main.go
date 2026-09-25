package indexeddb

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type IDBDatabase interface {
	CreateObjectStore(name string, keyPath string) IDBObjectStore
}

type IDBObjectStore interface {
}

func Open(name string, version int, onUpgradeNeeded func(db IDBDatabase)) (IDBDatabase, error) {
	resultCh := make(chan openResult, 1)
	request := app.Window().Get("indexedDB").Call("open", name, version)

	request.Call("addEventListener", "error", app.FuncOf(func(this app.Value, args []app.Value) any {
		errMsg := args[0].Get("target").Get("error").Get("message").String()
		resultCh <- openResult{db: nil, err: fmt.Errorf("IndexedDB error: %s", errMsg)}
		return nil
	}))

	request.Call("addEventListener", "success", app.FuncOf(func(this app.Value, args []app.Value) any {
		db := args[0].Get("target").Get("result")
		resultCh <- openResult{db: &idbDatabase{value: db}, err: nil}
		return nil
	}))

	request.Call("addEventListener", "upgradeneeded", app.FuncOf(func(this app.Value, args []app.Value) any {
		db := args[0].Get("target").Get("result")
		onUpgradeNeeded(&idbDatabase{value: db})
		return nil
	}))

	r := <-resultCh
	return r.db, r.err
}

type openResult struct {
	db  IDBDatabase
	err error
}
