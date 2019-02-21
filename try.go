// Package try is a Try/Catch implementation. A common pattern when
// needing to protect a section of code from panics, such as calling
// plugins or passing errors from deep in a recursive tree, is to use
// defer func() { recover() ... }() this is more or less the try catch
// pattern from other languages. This implementation generalizes that
// pattern and makes it easier to use. Unlike other languages, this turns
// a call into a value and a go error. Any caught value is transformed to
// an error. If it is already an error then it is passed along as is,
// otherwise fmt.Errorf is used to create an error. If one panics within
// an exeception this will become the returned error.
package try

import (
	"fmt"
	"reflect"

	"jsouthworth.net/go/dyn"
)

// New builds a new try/catch context out of the provided exception handlers.
func New(handlers ...exceptionHandler) func(fn interface{}) (interface{}, error) {
	hs := exceptionHandlers{
		handlers: make(map[reflect.Type]interface{}),
	}
	for _, handler := range handlers {
		handler(&hs)
	}
	return func(fn interface{}) (out interface{}, err error) {
		defer func(finally interface{}) {
			if finally != nil {
				got, ferr := Try(dyn.Bind(finally, out))
				out = got
				if err == nil {
					err = ferr
				}
			}
		}(hs.finally)
		defer func(hs *exceptionHandlers) {
			r := recover()
			if r == nil {
				return
			}
			rt := reflect.TypeOf(r)
			handler, ok := hs.handlers[rt]
			if ok {
				out, err = Try(dyn.Bind(handler, r))
				return
			}
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}(&hs)
		out = dyn.Apply(fn)
		return
	}
}

// Try builds a context from the try/catch handlers and then calls the
// provided function in that context.
func Try(
	fn interface{},
	handlers ...exceptionHandler,
) (out interface{}, err error) {
	return New(handlers...)(fn)
}

// Apply provides a simple interface for calling arbitrary functions,
// recovering from any panic, and finally returning the output and any
// error that occured.
func Apply(fn interface{}, args ...interface{}) (out interface{}, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("%v", v)
		}
	}()
	out = dyn.Apply(fn, args...)
	return
}

// Catch defines an exception handler. Fn must be a function of one
// argument. If multiple handlers with the same type are registered, the
// last one defined will be used.
func Catch(fn interface{}) exceptionHandler {
	return func(hs *exceptionHandlers) {
		fnt := reflect.TypeOf(fn)
		if fnt.Kind() != reflect.Func {
			return
		}
		if fnt.NumIn() != 1 {
			return
		}
		hs.handlers[fnt.In(0)] = fn
	}
}

// Finally defines what should happen after every call but before
// returning the value, error pair. This gives one final place to return
// a value regardless of the error condition. fn is a function of the type
// fn(cur rT, err error) fT where rT is the return type of the
// function/handlers and fT is the return type of finally.
func Finally(fn interface{}) exceptionHandler {
	return func(hs *exceptionHandlers) {
		hs.finally = fn
	}
}

type exceptionHandler func(*exceptionHandlers)

type exceptionHandlers struct {
	handlers map[reflect.Type]interface{}
	finally  interface{}
}
