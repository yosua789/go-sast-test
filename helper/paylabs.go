package helper

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func IsValidPaylabsRequest(ctx *gin.Context, path, payload, publicKey string) (res bool) {
	timestamp := ctx.GetHeader("X-TIMESTAMP")
	signature := ctx.GetHeader("X-SIGNATURE")

	// Decode base64 signature
	binarySignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode base64 signature")
		return false
	}

	// Compute SHA-256 hash of dataToSign
	hash := sha256.Sum256([]byte(payload))
	shaJson := fmt.Sprintf("%x", hash)
	signatureAfter := fmt.Sprintf("POST:%s:%s:%s", path, shaJson, timestamp)
	fmt.Println(signatureAfter)

	// Parse the public key
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Error().Msg("Failed to parse public key PEM block")
		return false
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse public key")
		return false
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		log.Error().Msg("Public key is not of type RSA")
		return false
	}

	// Verify the signature
	hashed := sha256.Sum256([]byte(signatureAfter))
	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed[:], binarySignature)
	if err != nil {
		log.Error().Err(err).Msg("Signature verification failed")
		return false
	}
	log.Info().Msg("Signature is valid.")
	return true
}

func GenerateSnapSignature(shaJson [32]byte, date, privateKeyPEM string) string {
	//  Parse the private key
	blockPrivate, _ := pem.Decode([]byte(privateKeyPEM))
	privateKey, err := x509.ParsePKCS1PrivateKey(blockPrivate.Bytes)
	if err != nil {
		panic(err)
	}
	// Generate the signature
	rawSignature := fmt.Sprintf("POST:/transfer-va/create-va:%x:%s", shaJson, date)
	h := sha256.New()
	h.Write([]byte(rawSignature))
	hashed := h.Sum(nil)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		panic(err)
	}

	// Base64 encode the signature
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	return signatureB64
}
