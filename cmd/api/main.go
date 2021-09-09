package main

import "github.com/nebisin/goExpense/internal/app"

func main() {
	s := app.NewServer()

	s.Run()
}
