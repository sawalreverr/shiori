package main

import (
	"fmt"
	"shiori/internal/config"
)

func main() {
	// load config
	cfg := config.DefaultConfig()

	// create store
	// newsStore := store.NewStore()

	fmt.Printf("%v", cfg)
}
