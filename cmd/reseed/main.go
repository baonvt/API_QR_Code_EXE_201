package main

import (
	"fmt"
	"go-api/config"
)

func main() {
	fmt.Println("🚀 Starting package reseed...")

	// Connect to database
	config.ConnectDatabase()

	// Run reseed
	if err := config.ReseedPackages(); err != nil {
		fmt.Printf("❌ Failed to reseed packages: %v\n", err)
		return
	}

	fmt.Println("✅ Packages reseeded successfully!")
}
