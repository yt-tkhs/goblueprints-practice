package main

import (
	"sync"
	"net/http"
	"fmt"
)

var (
	varsLock sync.RWMutex
	vars map[*http.Request]map[string]interface{}
)

func OpenVars(r *http.Request) {
	fmt.Println("Open: ", r)
	varsLock.Lock()
	if vars == nil {
		vars = map[*http.Request]map[string]interface{}{}
	}
	vars[r] = map[string]interface{}{}
	varsLock.Unlock()
}

func CloseVars(r *http.Request) {
	varsLock.Lock()
	delete(vars, r)
	varsLock.Unlock()
}

func GetVar(r *http.Request, key string) interface{} {
	//RLock(): 書き込みが行われていなければブロックされない = 複数の読み出しが同時に行える
	varsLock.RLock()
	value := vars[r][key]
	varsLock.RUnlock()
	return value
}

func SetVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	vars[r][key] = value
	varsLock.Unlock()
}
