package try

import (
	"fmt"

	"jsouthworth.net/go/dyn"
)

func ExampleTry() {
	out, err := Try(dyn.Bind(func(x int) int { panic("help!") }, 10))
	fmt.Println(out, err)
	// Output: <nil> help!
}

func ExampleTry_catch() {
	out, err := Try(dyn.Bind(func(x int) int { panic("help!") }, 10),
		Catch(func(s string) string { return s }))
	fmt.Println(out, err)
	// Output: help! <nil>
}

func ExampleTry_finally() {
	out, err := Try(dyn.Bind(func(x int) int { panic("help!") }, 10),
		Finally(func() string { return "Finally" }))
	fmt.Println(out, err)
	// Output: Finally help!
}

func ExampleTry_catchFinally() {
	out, err := Try(dyn.Bind(func(x int) int { panic("help!") }, 10),
		Catch(func(s string) string { return s }),
		Finally(func() { fmt.Println("Finally") }))
	fmt.Println(out, err)
	// Output: Finally
	// help! <nil>
}

func ExampleTry_catchFinallyNoResult() {
	out, err := Try(dyn.Bind(func(x int) int { panic("help!") }, 10),
		Catch(func(s string) {}),
		Finally(func() string { return "Finally" }))
	fmt.Println(out, err)
	// Output: Finally <nil>
}

func ExampleTry_noError() {
	out, err := Try(dyn.Bind(func(x int) int { return x }, 10),
		Catch(func(s string) interface{} { return nil }),
		Finally(func() string { return "Finally" }))
	fmt.Println(out, err)
	// Output: 10 <nil>
}

func ExampleTry_sendDoesNotUnderstand() {
	rcvr := &receiver{}
	out, err := Try(dyn.Bind(dyn.Send, rcvr, "Foo"),
		Catch(func(e dyn.ErrDoesNotUnderstand) interface{} {
			fmt.Println(e)
			return dyn.Send(rcvr, "String")
		}))
	fmt.Println(out, err)
	// Output: Object rcvr! does not understand [Foo]
	// rcvr! <nil>
}

func ExampleTry_origErrorPreserved() {
	rcvr := &receiver{}
	_, err := Try(dyn.Bind(dyn.Send, rcvr, "Foo"))
	fmt.Printf("%T\n", err)
	// Output: dyn.ErrDoesNotUnderstand
}

type receiver struct {
}

func (r *receiver) String() string {
	return "rcvr!"
}
