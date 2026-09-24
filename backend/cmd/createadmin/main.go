// Command createadmin provisions or resets one admin dashboard account.
// There's no self-serve signup for admin_users — this is the only way
// an account gets created, run by whoever already has shell/DB access:
//
//	go run ./cmd/createadmin -email you@perchly.app -password 'a strong password'
package main

import (
	"context"
	"flag"
	"log"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/config"
	"perchly-backend/internal/repository"
)

func main() {
	email := flag.String("email", "", "admin email (required)")
	password := flag.String("password", "", "admin password (required); re-running with an existing email resets it")
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("both -email and -password are required")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := repository.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	passwordHash, err := auth.HashAdminPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	admin, err := repository.NewAdminUserRepository(pool).Upsert(ctx, *email, passwordHash)
	if err != nil {
		log.Fatalf("create admin: %v", err)
	}

	log.Printf("admin ready: %s (%s)", admin.Email, admin.ID)
}
