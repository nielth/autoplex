package services

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
)

// func init() {
// 	// Example hardcoded private key seed (32 bytes for Ed25519 seed)
// 	privateKeySeedHex := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
// 	privateKeySeed, err := hex.DecodeString(privateKeySeedHex)
// 	if err != nil {
// 		log.Fatalf("Error decoding private key seed: %v", err)
// 	}
//
// 	// Ensure the length is 32 bytes for the seed
// 	if len(privateKeySeed) != ed25519.SeedSize {
// 		log.Fatalf("Invalid private key seed length: got %d, expected %d", len(privateKeySeed), ed25519.SeedSize)
// 	}
//
// 	// Generate the private key from the seed
// 	privateKey = ed25519.NewKeyFromSeed(privateKeySeed)
// 	publicKey = privateKey.Public().(ed25519.PublicKey)
// }

func init() {
	// Generate Ed25519 key pair dynamically
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Error generating Ed25519 key pair: " + err.Error())
	}
	privateKey = priv
	publicKey = pub
}

func CreateToken(username string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"username": username,
		// "exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat": time.Now().Unix(),
	})

	return claims.SignedString(privateKey)
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is EdDSA
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
