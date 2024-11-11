package main

import (
	"context"
	"fmt"
	"goeventbus"
	"log"
	"time"
)

type OrderCreatedEventRetry struct {
	OrderID string
}

func main() {
	bus := goeventbus.NewEventBus()

	// Suscribirse al evento OrderCreatedEvent con opciones avanzadas
	bus.Subscribe(OrderCreatedEventRetry{}, func(ctx context.Context, event interface{}) error {
		orderEvent := event.(OrderCreatedEventRetry)
		fmt.Printf("Processing order: %s\n", orderEvent.OrderID)
		return nil
	}, goeventbus.WithRetry(3, 2*time.Second), // Reintentos con backoff
		goeventbus.WithPriority(10), // Alta prioridad
		goeventbus.WithEventFilter(func(event interface{}) bool {
			return event.(OrderCreatedEventRetry).OrderID != ""
		}), // Filtrar eventos con OrderID no vacío
		goeventbus.WithErrorLog(func(err error) {
			if err != nil {
				log.Printf("Error processing event: %v", err)
			}
		}),
	)

	// Publicar un evento
	bus.PublishEvent(context.Background(), OrderCreatedEventRetry{OrderID: "order-123"})

	// Esperar para permitir que el evento se procese
	time.Sleep(5 * time.Second)

	// Cerrar el EventBus
	bus.Close()
}
