package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	// Command line flags
	userID := flag.String("user", "test-user-001", "User ID for the token")
	email := flag.String("email", "test@hexabase.local", "Email for the token")
	name := flag.String("name", "Test User", "Display name for the token")
	orgID := flag.String("org", "", "Organization ID (optional)")
	expHours := flag.Int("exp", 24, "Token expiration in hours")
	pretty := flag.Bool("pretty", false, "Pretty print the token claims")
	
	flag.Parse()

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Prepare organization IDs
	orgIDs := []string{}
	if *orgID != "" {
		orgIDs = append(orgIDs, *orgID)
	}

	// Create claims
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": *userID,
		"email":   *email,
		"name":    *name,
		"org_ids": orgIDs,
		"iss":     "https://api.hexabase.local",
		"sub":     *userID,
		"aud":     []string{"hexabase-api"},
		"exp":     now.Add(time.Duration(*expHours) * time.Hour).Unix(),
		"nbf":     now.Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.New().String(),
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	// Output
	if *pretty {
		fmt.Println("Token Claims:")
		fmt.Println("=============")
		claimsJSON, _ := json.MarshalIndent(claims, "", "  ")
		fmt.Println(string(claimsJSON))
		fmt.Println("\nToken:")
		fmt.Println("======")
	}
	
	fmt.Printf("Bearer %s\n", tokenString)
	
	if !*pretty {
		fmt.Println("\nTo use this token:")
		fmt.Println("export TOKEN=\"Bearer " + tokenString[:50] + "...\"")
		fmt.Println("\nExample usage:")
		fmt.Println("curl http://localhost:8080/api/v1/organizations -H \"Authorization: $TOKEN\"")
	}
}

// To install dependencies:
// go get github.com/golang-jwt/jwt/v5
// go get github.com/google/uuid

// Usage examples:
// go run generate_test_token.go
// go run generate_test_token.go -user test-user-001 -email test@hexabase.local -org test-org-001
// go run generate_test_token.go -pretty