package main

import (
	"context"
	"fmt"
	"goeventbus"
	"log"
	"time"
)

type OrderCreatedEventb struct {
	OrderID string
}

func logMiddleware(ctx context.Context, event interface{}, next func(ctx context.Context, event interface{}) error) error {
	start := time.Now()
	err := next(ctx, event)
	fmt.Printf("Event processed in: %v\n", time.Since(start))
	return err
}

func main() {
	bus := goeventbus.NewEventBus(5) // Limita a 5 goroutines concurrentes

	bus.Subscribe(OrderCreatedEventb{}, func(ctx context.Context, event interface{}) error {
		order := event.(OrderCreatedEventb)
		time.Sleep(1 * time.Second) // Simula procesamiento
		fmt.Printf("Order created and processed: %s\n", order.OrderID)
		return nil
	}, goeventbus.WithErrorLog(func(err error) {
		if err != nil {
			log.Printf("Error processing event: %v\n", err)
		}
	}), goeventbus.WithMiddleware(logMiddleware))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	bus.PublishEvent(ctx, OrderCreatedEventb{OrderID: "order-123"})
	time.Sleep(2 * time.Second)
	bus.Close()
}
