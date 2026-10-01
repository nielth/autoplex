package services

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
)

// InitJWTKeys sets the Ed25519 signing key. With JWT_SECRET set, the key is
// derived from it, so logins survive restarts and redeploys. Without it a
// random key is used and everyone is logged out on every restart. It must run
// after the .env file is loaded.
func InitJWTKeys() error {
	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		seed := sha256.Sum256([]byte(secret))
		privateKey = ed25519.NewKeyFromSeed(seed[:])
		publicKey = privateKey.Public().(ed25519.PublicKey)
		return nil
	}

	log.Printf("JWT_SECRET is not set, using a random key: logins will not survive a restart")
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("error generating Ed25519 key pair: %w", err)
	}
	privateKey = priv
	publicKey = pub
	return nil
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
