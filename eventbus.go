package goeventbus

import (
	"context"
	"errors"
	"log"
	"reflect"
	"sync"
)

// EventBus es la interfaz para nuestro bus de eventos, definiendo los métodos básicos
type EventBus interface {
	PublishEvent(ctx context.Context, event interface{}) error
	Subscribe(eventType interface{}, handler func(ctx context.Context, event interface{}) error, options ...ListenerOption) error
	Unsubscribe(eventType interface{}, handler func(ctx context.Context, event interface{}) error) error
	Close()
}

// ConcreteEventBus es la implementación del EventBus
type ConcreteEventBus struct {
	subscribers   map[reflect.Type][]*Listener
	lock          sync.RWMutex
	maxGoroutines int
	sem           chan struct{}
}

// Listener representa un suscriptor de eventos con opciones de configuración
type Listener struct {
	handler    func(ctx context.Context, event interface{}) error
	errorLog   func(error)
	middleware []Middleware
}

// Middleware define una función que puede ejecutarse antes o después del handler
type Middleware func(ctx context.Context, event interface{}, next func(ctx context.Context, event interface{}) error) error

// ListenerOption define opciones configurables para un Listener
type ListenerOption func(*Listener)

// WithErrorLog permite especificar una función para registrar errores
func WithErrorLog(logFn func(error)) ListenerOption {
	return func(l *Listener) {
		l.errorLog = logFn
	}
}

// WithMiddleware permite especificar un middleware para el listener
func WithMiddleware(mw Middleware) ListenerOption {
	return func(l *Listener) {
		l.middleware = append(l.middleware, mw)
	}
}

// DefaultMaxGoroutines es el valor predeterminado de goroutines concurrentes si no se proporciona uno
const DefaultMaxGoroutines = 100

// NewEventBus crea una nueva instancia de ConcreteEventBus con maxGoroutines opcional
func NewEventBus(maxGoroutines ...int) *ConcreteEventBus {
	goroutines := DefaultMaxGoroutines
	if len(maxGoroutines) > 0 && maxGoroutines[0] > 0 {
		goroutines = maxGoroutines[0]
	}

	return &ConcreteEventBus{
		subscribers:   make(map[reflect.Type][]*Listener),
		maxGoroutines: goroutines,
		sem:           make(chan struct{}, goroutines),
	}
}

// PublishEvent envía un evento a todos los listeners suscritos de ese tipo, usando el contexto proporcionado
func (bus *ConcreteEventBus) PublishEvent(ctx context.Context, event interface{}) error {
	eventType := reflect.TypeOf(event)
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	if listeners, found := bus.subscribers[eventType]; found {
		for _, listener := range listeners {
			bus.sem <- struct{}{} // Limita el número de goroutines simultáneas
			go func(listener *Listener) {
				defer func() { <-bus.sem }() // Libera el espacio en el semáforo
				err := bus.invokeListener(ctx, listener, event)
				if err != nil && listener.errorLog != nil {
					listener.errorLog(err)
				}
			}(listener)
		}
	}
	return nil
}

// Subscribe se suscribe a un tipo de evento específico con opciones de configuración
func (bus *ConcreteEventBus) Subscribe(eventType interface{}, handler func(ctx context.Context, event interface{}) error, options ...ListenerOption) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	listener := &Listener{
		handler: handler,
	}

	// Aplicar opciones configurables
	for _, option := range options {
		option(listener)
	}

	eventTypeKey := reflect.TypeOf(eventType)
	bus.subscribers[eventTypeKey] = append(bus.subscribers[eventTypeKey], listener)
	return nil
}

// Unsubscribe elimina un listener del bus de eventos
func (bus *ConcreteEventBus) Unsubscribe(eventType interface{}, handler func(ctx context.Context, event interface{}) error) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	eventTypeKey := reflect.TypeOf(eventType)
	if listeners, found := bus.subscribers[eventTypeKey]; found {
		for i, listener := range listeners {
			if reflect.ValueOf(listener.handler).Pointer() == reflect.ValueOf(handler).Pointer() {
				// Eliminar listener
				bus.subscribers[eventTypeKey] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
	return nil
}

// Close cierra todos los listeners y libera los recursos
func (bus *ConcreteEventBus) Close() {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	for _, listeners := range bus.subscribers {
		for _, listener := range listeners {
			if listener.errorLog != nil {
				listener.errorLog(nil)
			}
		}
	}
}

// invokeListener ejecuta el handler del listener con soporte para contextos y middleware
func (bus *ConcreteEventBus) invokeListener(ctx context.Context, listener *Listener, event interface{}) error {
	if len(listener.middleware) == 0 {
		return listener.handler(ctx, event)
	}

	// Ejecutar middleware en cadena
	var current int
	var next func(ctx context.Context, event interface{}) error
	next = func(ctx context.Context, event interface{}) error {
		if current >= len(listener.middleware) {
			return listener.handler(ctx, event)
		}
		mw := listener.middleware[current]
		current++
		return mw(ctx, event, next)
	}
	return next(ctx, event)
}

// logError es una función auxiliar para crear un error de log
func logError(message string) error {
	log.Println(message)
	return errors.New(message)
}
