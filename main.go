package main

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits  = "0123456789"
	charsetSymbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

type GenerateRequest struct {
	Length      int  `json:"length" form:"length"`
	UseLower    bool `json:"use_lower" form:"use_lower"`
	UseUpper    bool `json:"use_upper" form:"use_upper"`
	UseDigits   bool `json:"use_digits" form:"use_digits"`
	UseSymbols  bool `json:"use_symbols" form:"use_symbols"`
}

type GenerateResponse struct {
	Password string `json:"password"`
	Length   int    `json:"length"`
}

func generatePassword(req GenerateRequest) (string, error) {
	charset := ""
	if req.UseLower {
		charset += charsetLower
	}
	if req.UseUpper {
		charset += charsetUpper
	}
	if req.UseDigits {
		charset += charsetDigits
	}
	if req.UseSymbols {
		charset += charsetSymbols
	}
	if charset == "" {
		charset = charsetLower + charsetUpper + charsetDigits
	}

	password := make([]byte, req.Length)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[n.Int64()]
	}
	return string(password), nil
}

func handleGenerate(c *gin.Context) {
	req := GenerateRequest{
		Length:    16,
		UseLower:  true,
		UseUpper:  true,
		UseDigits: true,
	}

	// クエリパラメータを優先的に使用
	if l := c.Query("length"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 128 {
			req.Length = n
		}
	}
	if v := c.Query("use_lower"); v != "" {
		req.UseLower = v == "true" || v == "1"
	}
	if v := c.Query("use_upper"); v != "" {
		req.UseUpper = v == "true" || v == "1"
	}
	if v := c.Query("use_digits"); v != "" {
		req.UseDigits = v == "true" || v == "1"
	}
	if v := c.Query("use_symbols"); v != "" {
		req.UseSymbols = v == "true" || v == "1"
	}

	// POSTの場合はJSONボディも受け付ける
	if c.Request.Method == http.MethodPost {
		c.ShouldBindJSON(&req)
	}

	if req.Length <= 0 || req.Length > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "length must be between 1 and 128"})
		return
	}

	password, err := generatePassword(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate password"})
		return
	}

	c.JSON(http.StatusOK, GenerateResponse{
		Password: password,
		Length:   req.Length,
	})
}

func main() {
	r := gin.Default()

	r.GET("/generate", handleGenerate)
	r.POST("/generate", handleGenerate)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
