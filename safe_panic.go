package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

var safePanicOnce sync.Once

func safePanic(r any, s tcell.Screen) {
	safePanicOnce.Do(func() {
		s.Fini()
		if r != nil {
			panic(r)
		}
	})
}

func safeGo(fn func(), s tcell.Screen) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				safePanic(r, s)
			}
		}()
		fn()
	}()
}
