package goeventbus_test

import (
	"context"
	"fmt"
	"goeventbus"
	"sync"
	"testing"
	"time"
)

// Evento de prueba
type OrderCreatedEvent struct {
	OrderID string
}

// Test para verificar la publicación y suscripción básica de eventos
func TestPublishAndSubscribe(t *testing.T) {
	bus := goeventbus.NewEventBus()

	// Variable para verificar si el listener fue llamado
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		orderEvent := event.(OrderCreatedEvent)
		if orderEvent.OrderID != "order-123" {
			t.Errorf("Unexpected OrderID, got %s, want %s", orderEvent.OrderID, "order-123")
		}
		wg.Done()
		return nil
	})

	// Publicar evento
	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})
	wg.Wait() // Esperar a que el listener sea llamado

	bus.Close()
}

// Test para verificar la funcionalidad de reintento con backoff
func TestRetryWithBackoff(t *testing.T) {
	bus := goeventbus.NewEventBus()
	var attemptCount int

	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		attemptCount++
		return fmt.Errorf("failed") // Simula un error para activar el retry
	}, goeventbus.WithRetry(3, 100*time.Millisecond))

	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})

	time.Sleep(500 * time.Millisecond) // Esperar para que los reintentos se completen

	if attemptCount != 4 { // 1 intento inicial + 3 reintentos
		t.Errorf("Expected 4 attempts, got %d", attemptCount)
	}

	bus.Close()
}

// Test para verificar que los listeners con mayor prioridad se ejecutan primero
func TestListenerPriority(t *testing.T) {
	bus := goeventbus.NewEventBus()
	var executionOrder []int

	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		executionOrder = append(executionOrder, 1)
		return nil
	}, goeventbus.WithPriority(1))

	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		executionOrder = append(executionOrder, 2)
		return nil
	}, goeventbus.WithPriority(10)) // Mayor prioridad

	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})

	time.Sleep(100 * time.Millisecond) // Dar tiempo para que los listeners se ejecuten

	if len(executionOrder) != 2 || executionOrder[0] != 2 || executionOrder[1] != 1 {
		t.Errorf("Expected execution order [2, 1], got %v", executionOrder)
	}

	bus.Close()
}

// Test para verificar que el filtro de eventos funcione correctamente
func TestEventFilter(t *testing.T) {
	bus := goeventbus.NewEventBus()
	var eventProcessed bool

	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		eventProcessed = true
		return nil
	}, goeventbus.WithEventFilter(func(event interface{}) bool {
		// Solo procesar eventos con OrderID igual a "order-123"
		return event.(OrderCreatedEvent).OrderID == "order-123"
	}))

	// Publicar evento que no cumple con el filtro
	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-456"})

	time.Sleep(100 * time.Millisecond) // Dar tiempo para la ejecución del listener

	if eventProcessed {
		t.Errorf("Event was processed despite not meeting filter criteria")
	}

	// Publicar evento que cumple con el filtro
	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})

	time.Sleep(100 * time.Millisecond) // Dar tiempo para la ejecución del listener

	if !eventProcessed {
		t.Errorf("Event was not processed despite meeting filter criteria")
	}

	bus.Close()
}

// Test para verificar el manejo de errores en el log
func TestErrorLog(t *testing.T) {
	bus := goeventbus.NewEventBus()

	var errorLogged bool
	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		return fmt.Errorf("intentional error")
	}, goeventbus.WithErrorLog(func(err error) {
		if err != nil && err.Error() == "intentional error" {
			errorLogged = true
		}
	}))

	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})

	time.Sleep(100 * time.Millisecond) // Dar tiempo para la ejecución del listener

	if !errorLogged {
		t.Errorf("Expected error was not logged")
	}

	bus.Close()
}
