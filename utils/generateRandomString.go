package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	mathRand "math/rand"
	"time"
)

// GenerateSecureToken génère un token sécurisé pour la réinitialisation de mot de passe
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomString génère une chaîne aléatoire pour des usages non-critiques
func GenerateRandomString(length int) string {
	// Initialisation du seed avec l'heure actuelle
	mathRand.Seed(time.Now().UnixNano())

	var charSet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		bytes[i] = charSet[n.Int64()]
	}
	return string(bytes)
}
