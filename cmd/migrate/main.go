package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"kisanlink-ecom/migrations"
)

func main() {
	var (
		seedRBAC = flag.Bool("seed-rbac", false, "Seed e-commerce RBAC data to AAA service")
	)
	flag.Parse()

	if *seedRBAC {
		// Run RBAC seeding
		if err := migrations.SeedEcommerceRBAC(); err != nil {
			log.Printf("RBAC seeding failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("RBAC seeding completed successfully!")
		return
	}

	// Default behavior - just show help
	fmt.Println("Use --seed-rbac to seed e-commerce RBAC data to AAA service")
	fmt.Println("Database migrations are handled separately")
}
