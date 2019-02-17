// Package try is a Try/Catch language extension
package try

import (
	"fmt"
	"reflect"

	"jsouthworth.net/go/dyn"
)

func Try(
	fn interface{},
	handlers ...exceptionHandler,
) (out interface{}, err error) {
	hs := exceptionHandlers{
		handlers: make(map[reflect.Type]interface{}),
	}
	for _, handler := range handlers {
		handler(&hs)
	}
	defer func(finally interface{}) {
		if finally != nil {
			got, ferr := Try(finally)
			if out == nil {
				out = got
			}
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
