package main

import (
	"fmt"

	"github.com/rbrabson/dgame/action"
)

type Foo struct {
	action.ActionBase
}

func main() {
	var action action.Action
	action = &Foo{}
	action.Execute()
	fmt.Println(action.IsFinished())
}

func (f *Foo) Execute() {
	fmt.Println("My Foo")
}
