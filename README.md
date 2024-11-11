# goeventbus

`goeventbus` is an event publishing and subscribing library in Go, designed to facilitate asynchronous communication within applications. It allows developers to publish events and listen for specific events in different parts of the application.

## Installation

```bash
go get -u github.com/ANDERSON1808/goeventbus
```

## Features

- Publish and subscribe to events within your application.
- Configurable retries and backoff in case of processing failures.
- Listener priority handling to control the order of execution.
- Event filtering to ensure listeners only process relevant events.
- Optional middleware support to apply additional processing logic.

## Basic Usage

### Creating an Event and Setting Up the EventBus

Define an event type, initialize an `EventBus`, and subscribe to listen for the event.

```go
package main

import (
	"context"
	"fmt"
	"goeventbus"
	"time"
)

type OrderCreatedEvent struct {
	OrderID string
}

func main() {
	// Initialize the EventBus
	bus := goeventbus.NewEventBus()

	// Subscribe to the OrderCreatedEvent
	bus.Subscribe(OrderCreatedEvent{}, func(ctx context.Context, event interface{}) error {
		order := event.(OrderCreatedEvent)
		fmt.Printf("Processing order: %s\n", order.OrderID)
		return nil
	})

	// Publish an event
	bus.PublishEvent(context.Background(), OrderCreatedEvent{OrderID: "order-123"})

	// Allow time for the event to be processed
	time.Sleep(1 * time.Second)

	// Close the EventBus
	bus.Close()
}
```

## Advanced Features

### Retry with Backoff

You can configure a listener to retry a specified number of times with a delay (backoff) between retries.

```go
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

	// Subscribe with retry and backoff settings
	bus.Subscribe(OrderCreatedEventRetry{}, func(ctx context.Context, event interface{}) error {
		order := event.(OrderCreatedEventRetry)
		fmt.Printf("Processing order: %s\n", order.OrderID)
		return fmt.Errorf("simulated error") // Simulate an error for retries
	}, goeventbus.WithRetry(3, 500*time.Millisecond), // 3 retries with 500ms backoff
		goeventbus.WithErrorLog(func(err error) {
			log.Printf("Error processing event: %v", err)
		}),
	)

	// Publish an event
	bus.PublishEvent(context.Background(), OrderCreatedEventRetry{OrderID: "order-123"})

	// Allow time for the retries to occur
	time.Sleep(3 * time.Second)

	bus.Close()
}
```

### Prioritized Listeners

Listeners can be given a priority level to control the order of execution. Higher priority listeners are executed before lower priority ones.

```go
package main

import (
	"context"
	"fmt"
	"goeventbus"
	"time"
)

type OrderCreatedEventPriority struct {
	OrderID string
}

func main() {
	bus := goeventbus.NewEventBus()

	// Subscribe with low priority
	bus.Subscribe(OrderCreatedEventPriority{}, func(ctx context.Context, event interface{}) error {
		fmt.Println("Low priority listener")
		return nil
	}, goeventbus.WithPriority(1))

	// Subscribe with high priority
	bus.Subscribe(OrderCreatedEventPriority{}, func(ctx context.Context, event interface{}) error {
		fmt.Println("High priority listener")
		return nil
	}, goeventbus.WithPriority(10))

	// Publish an event
	bus.PublishEvent(context.Background(), OrderCreatedEventPriority{OrderID: "order-456"})

	// Allow time for listeners to execute
	time.Sleep(1 * time.Second)

	bus.Close()
}
```

### Event Filtering

You can apply a filter to listeners so they only process events that meet specific conditions.

```go
package main

import (
	"context"
	"fmt"
	"goeventbus"
	"time"
)

type OrderCreatedEventFiltered struct {
	OrderID string
}

func main() {
	bus := goeventbus.NewEventBus()

	// Subscribe with a filter
	bus.Subscribe(OrderCreatedEventFiltered{}, func(ctx context.Context, event interface{}) error {
		order := event.(OrderCreatedEventFiltered)
		fmt.Printf("Filtered order processed: %s\n", order.OrderID)
		return nil
	}, goeventbus.WithEventFilter(func(event interface{}) bool {
		// Only process orders with a specific OrderID
		return event.(OrderCreatedEventFiltered).OrderID == "order-123"
	}))

	// Publish events
	bus.PublishEvent(context.Background(), OrderCreatedEventFiltered{OrderID: "order-123"}) // This will be processed
	bus.PublishEvent(context.Background(), OrderCreatedEventFiltered{OrderID: "order-456"}) // This will be ignored

	// Allow time for listeners to execute
	time.Sleep(1 * time.Second)

	bus.Close()
}
```

## Closing the EventBus

It’s important to close the EventBus to release resources when you are done:

```go
bus.Close()
```

## License

MIT License
```

### Summary of Examples

- **Basic Usage**: Publish and subscribe to events.
- **Retry with Backoff**: Add automatic retry for events with a delay between attempts.
- **Prioritized Listeners**: Control the order of listener execution based on priority.
- **Event Filtering**: Ensure only specific events are processed by a listener.

This README provides comprehensive usage instructions and examples for `goeventbus`, making it easier to integrate into various applications.