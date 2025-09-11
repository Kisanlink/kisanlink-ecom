package migrations

import (
    "fmt"
)

// SeedEcommerceRBAC seeds the AAA service with e-commerce specific RBAC data
// TODO: Re-enable when AAA service is properly configured
func SeedEcommerceRBAC() error {
    fmt.Println("RBAC seeding is currently disabled - AAA service configuration needed")
    return nil
}
