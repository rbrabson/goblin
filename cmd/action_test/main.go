package main

import (
	"fmt"

	"github.com/rbrabson/dgame/action"
)

type Foo struct {
	action.ActionBase
	str string
}

func main() {
	action := &Foo{str: "Execute Foo"}
	action.Initialize()
	action.Execute()
	fmt.Println(action.IsFinished())
	fmt.Printf("Main: %s\n", action)
}

func (f *Foo) Execute() {
	fmt.Println(f.str)
}

func (f *Foo) String() string {
	return f.str
}
