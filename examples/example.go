package main

import (
	"context"
	"fmt"
	"goeventbus"
	"log"
	"time"
)

type OrderCreatedEvent struct {
	OrderID string
}

func main() {
	bus := goeventbus.NewEventBus()

	// Listener para OrderCreatedEvent con registro de errores
	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		order := event.(OrderCreatedEvent)
		select {
		case <-time.After(3 * time.Second): // Simula un procesamiento largo
			fmt.Printf("Processing order: %s\n", order.OrderID)
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}, goeventbus.WithErrorLog(func(err error) {
		if err != nil {
			log.Printf("Error processing event: %v\n", err)
		}
	}))

	// Publicar un evento con un contexto con límite de tiempo de 4 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	bus.PublishEvent(ctx, OrderCreatedEvent{OrderID: "order-123"})

	// Esperar tiempo suficiente para ver los resultados
	time.Sleep(5 * time.Second)

	// Cerrar el EventBus al finalizar
	bus.Close()
}
