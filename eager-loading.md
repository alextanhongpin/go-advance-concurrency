## Eager loading with errgroup

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

type UserAggregate struct {
	User  *User
	Order *Order
}

func (a *UserAggregate) Load(ctx context.Context, g *errgroup.Group, userID string) {
	g.Go(func() error { return a.fetchUser(ctx, userID) })
	g.Go(func() error {
		if rand.Intn(2) == 3 {
			time.Sleep(300 * time.Millisecond)
			return errors.New("failed")
		}
		if err := a.fetchOrder(ctx, userID); err != nil {
			return err
		}
		g.Go(func() error { return a.fetchAddress(ctx, a.Order.ID) })
		g.Go(func() error { return a.fetchShipment(ctx, a.Order.ID) })
		return nil
	})
}

func (a *UserAggregate) fetchUser(ctx context.Context, userID string) error {
	fmt.Println("fetching user", userID)
	select {
	case <-time.After(1 * time.Second):
		a.User = &User{ID: userID}
		return nil
	case <-ctx.Done():
		fmt.Println("aborting")
		return ctx.Err()
	}
}

func (a *UserAggregate) fetchOrder(ctx context.Context, userID string) error {
	fmt.Println("fetching order", userID)
	select {
	case <-time.After(1 * time.Second):
		a.Order = &Order{ID: "order-id", UserID: userID}
		return nil
	case <-ctx.Done():
		fmt.Println("aborting")
		return ctx.Err()
	}
}

func (a *UserAggregate) fetchAddress(ctx context.Context, orderID string) error {
	fmt.Println("fetching address", orderID)
	select {
	case <-time.After(1 * time.Second):
		a.Order.Address = &Address{ID: "address-id", OrderID: orderID}
		return nil
	case <-ctx.Done():
		fmt.Println("aborting")
		return ctx.Err()
	}
}

func (a *UserAggregate) fetchShipment(ctx context.Context, orderID string) error {
	fmt.Println("fetching shipment", orderID)
	select {
	case <-time.After(1 * time.Second):
		a.Order.Shipment = &Shipment{ID: "shipment-id", OrderID: orderID}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type User struct {
	ID string
}

type Order struct {
	ID       string
	UserID   string
	Address  *Address
	Shipment *Shipment
}
type Address struct {
	ID      string
	OrderID string
}

type Shipment struct {
	ID      string
	OrderID string
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	n := 10
	agg := make([]*UserAggregate, n)
	g, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < n; i++ {
		agg[i] = new(UserAggregate)
		agg[i].Load(ctx, g, fmt.Sprintf("user-%d", i))
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	for _, a := range agg {
		fmt.Println("received", a.User, a.Order)
	}
}
```
