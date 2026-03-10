package main

import (
	"github.com/guncv/ticket-reservation-server/internal/containers"
	_ "github.com/guncv/ticket-reservation-server/internal/handlers"
)

func main() {
	c := containers.NewContainer()
	if err := c.Run().Error; err != nil {
		panic(err)
	}
}
