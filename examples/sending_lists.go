package main

import (
	"context"
	"fmt"
	"goeventbus"
	"time"
)

// Product representa un producto en una orden
type Product struct {
	Name     string
	Quantity int
	Price    float64
}

// OrderCreatedEventList representa un evento para la creación de una orden
type OrderCreatedEventList struct {
	OrderID   string
	Customer  string
	Products  []Product
	TotalCost float64
}

func main() {
	bus := goeventbus.NewEventBus()

	// Suscribirse a eventos de tipo OrderCreatedEventList con un listener
	bus.Subscribe(OrderCreatedEventList{}, func(ctx context.Context, event interface{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err() // Maneja la cancelación o timeout
		default:
			orderEvent := event.(OrderCreatedEventList)
			fmt.Printf("Processing order: %s for customer: %s\n", orderEvent.OrderID, orderEvent.Customer)
			fmt.Println("Products in the order:")
			for _, product := range orderEvent.Products {
				fmt.Printf("- %s (x%d) at $%.2f each\n", product.Name, product.Quantity, product.Price)
			}
			fmt.Printf("Total Cost: $%.2f\n\n", orderEvent.TotalCost)
			return nil
		}
	})

	// Publicar un evento con datos complejos usando un contexto con límite de tiempo
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	order := OrderCreatedEventList{
		OrderID:  "order-789",
		Customer: "Alice Smith",
		Products: []Product{
			{Name: "Laptop", Quantity: 1, Price: 999.99},
			{Name: "Mouse", Quantity: 2, Price: 19.99},
			{Name: "Keyboard", Quantity: 1, Price: 49.99},
		},
		TotalCost: 1089.96,
	}

	bus.PublishEvent(ctx, order)

	// Dar tiempo para procesar todos los eventos antes de finalizar
	time.Sleep(2 * time.Second)

	// Cerrar el EventBus al terminar (esto cierra los canales)
	bus.Close()
}
