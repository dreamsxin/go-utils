package bus

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// HandlerFunc defines a handler function interface.
type HandlerFunc any

// Msg defines a message interface.
type Msg any

// ErrHandlerNotFound defines an error if a handler is not found.
var ErrHandlerNotFound = errors.New("handler not found")

// Bus type defines the bus interface structure.
type Bus interface {
	Publish(ctx context.Context, msg Msg) error
	AddEventListener(handler HandlerFunc)
}

// InProcBus defines the bus structure.
type InProcBus struct {
	listeners map[string][]HandlerFunc
}

func ProvideBus() *InProcBus {
	return &InProcBus{
		listeners: make(map[string][]HandlerFunc),
	}
}

// Publish function publish a message to the bus listener.
func (b *InProcBus) Publish(ctx context.Context, msg Msg) error {
	v := reflect.TypeOf(msg)
	msgName := ""
	if v.Kind() == reflect.Ptr {
		msgName = "p:" + v.Elem().Name()
	} else {
		msgName = v.Name()
	}

	var params = []reflect.Value{}
	if listeners, exists := b.listeners[msgName]; exists {
		params = append(params, reflect.ValueOf(ctx))
		params = append(params, reflect.ValueOf(msg))
		if err := callListeners(listeners, params); err != nil {
			return err
		}
	}

	return nil
}

func callListeners(listeners []HandlerFunc, params []reflect.Value) error {
	for _, listenerHandler := range listeners {
		ret := reflect.ValueOf(listenerHandler).Call(params)
		e := ret[0].Interface()
		if e != nil {
			err, ok := e.(error)
			if ok {
				return err
			}
			return fmt.Errorf("expected listener to return an error, got '%T'", e)
		}
	}
	return nil
}

func (b *InProcBus) AddEventListener(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	v := handlerType.In(1)
	eventName := ""
	if v.Kind() == reflect.Ptr {
		eventName = "p:" + v.Elem().Name()
	} else {
		eventName = v.Name()
	}
	_, exists := b.listeners[eventName]
	if !exists {
		b.listeners[eventName] = make([]HandlerFunc, 0)
	}
	b.listeners[eventName] = append(b.listeners[eventName], handler)
}
