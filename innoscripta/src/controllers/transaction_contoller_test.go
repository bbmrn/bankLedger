package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"INNOSCRIPTA/src/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProcessTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/transaction", ProcessTransaction)

	t.Run("Valid Transaction", func(t *testing.T) {
		transaction := models.Transaction{
			UserID: 123,
			Amount: 100,
		}
		jsonValue, _ := json.Marshal(transaction)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(jsonValue))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.Equal(t, "Transaction processed successfully", response["message"])
	})

	t.Run("Invalid Transaction", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer([]byte("invalid json")))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetTransactionHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/transactions/:id", GetTransactionHistory)

	t.Run("Valid User ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/transactions/123", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.Equal(t, "123", response["user_id"])
	})

	t.Run("Empty User ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/transactions/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
