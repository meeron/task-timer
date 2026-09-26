package indexeddb

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type idbObjectStore struct {
	value app.Value
}

func (s *idbObjectStore) Add(record map[string]interface{}) error {
	resultCh := make(chan error, 1)
	req := s.value.Call("add", record)
	req.Set("onerror", app.FuncOf(func(this app.Value, args []app.Value) any {
		msg := args[0].Get("target").Get("error").Get("message").String()
		resultCh <- fmt.Errorf("IndexedDB add error: %s", msg)
		return nil
	}))
	req.Set("onsuccess", app.FuncOf(func(this app.Value, args []app.Value) any {
		resultCh <- nil
		return nil
	}))
	return <-resultCh
}

func (s *idbObjectStore) Put(record map[string]interface{}) error {
	resultCh := make(chan error, 1)
	req := s.value.Call("put", record)
	req.Set("onerror", app.FuncOf(func(this app.Value, args []app.Value) any {
		msg := args[0].Get("target").Get("error").Get("message").String()
		resultCh <- fmt.Errorf("IndexedDB put error: %s", msg)
		return nil
	}))
	req.Set("onsuccess", app.FuncOf(func(this app.Value, args []app.Value) any {
		resultCh <- nil
		return nil
	}))
	return <-resultCh
}

func (s *idbObjectStore) Delete(key string) error {
	resultCh := make(chan error, 1)
	req := s.value.Call("delete", key)
	req.Set("onerror", app.FuncOf(func(this app.Value, args []app.Value) any {
		msg := args[0].Get("target").Get("error").Get("message").String()
		resultCh <- fmt.Errorf("IndexedDB delete error: %s", msg)
		return nil
	}))
	req.Set("onsuccess", app.FuncOf(func(this app.Value, args []app.Value) any {
		resultCh <- nil
		return nil
	}))
	return <-resultCh
}

type getAllResult struct {
	values []app.Value
	err    error
}

func (s *idbObjectStore) GetAll() ([]app.Value, error) {
	resultCh := make(chan getAllResult, 1)
	req := s.value.Call("getAll")
	req.Set("onerror", app.FuncOf(func(this app.Value, args []app.Value) any {
		msg := args[0].Get("target").Get("error").Get("message").String()
		resultCh <- getAllResult{err: fmt.Errorf("IndexedDB getAll error: %s", msg)}
		return nil
	}))
	req.Set("onsuccess", app.FuncOf(func(this app.Value, args []app.Value) any {
		jsArray := args[0].Get("target").Get("result")
		length := jsArray.Length()
		values := make([]app.Value, length)
		for i := 0; i < length; i++ {
			values[i] = jsArray.Index(i)
		}
		resultCh <- getAllResult{values: values}
		return nil
	}))
	r := <-resultCh
	return r.values, r.err
}
