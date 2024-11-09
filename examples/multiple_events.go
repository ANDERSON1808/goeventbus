package main

import (
	"context"
	"fmt"
	"goeventbus"
	"time"
)

// OrderCreatedEvents define un evento para la creación de una orden
type OrderCreatedEvents struct {
	OrderID string
}

// UserCreatedEvent define un evento para la creación de un usuario
type UserCreatedEvent struct {
	Username string
}

func main() {
	bus := goeventbus.NewEventBus()

	// Suscribirse a eventos de tipo OrderCreatedEvent
	bus.Subscribe(OrderCreatedEvents{}, func(ctx context.Context, event interface{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err() // Maneja la cancelación o timeout
		default:
			orderEvent := event.(OrderCreatedEvents)
			fmt.Printf("Order created and processed: %s\n", orderEvent.OrderID)
			return nil
		}
	})

	// Suscribirse a eventos de tipo UserCreatedEvent
	bus.Subscribe(UserCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err() // Maneja la cancelación o timeout
		default:
			userEvent := event.(UserCreatedEvent)
			fmt.Printf("User created and processed: %s\n", userEvent.Username)
			return nil
		}
	})

	// Publicar eventos de diferentes tipos usando un contexto con límite de tiempo
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		bus.PublishEvent(ctx, OrderCreatedEvents{OrderID: fmt.Sprintf("order-%d", i)})
		bus.PublishEvent(ctx, UserCreatedEvent{Username: fmt.Sprintf("user_%d", i)})

		time.Sleep(1 * time.Second) // Simula un retraso en la publicación de eventos
	}

	// Dar tiempo para procesar todos los eventos antes de finalizar
	time.Sleep(2 * time.Second)

	// Cerrar el EventBus al terminar (esto cierra los canales)
	bus.Close()
}
