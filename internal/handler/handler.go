package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Falokut/intern_test_task/internal/models"
	"github.com/Falokut/intern_test_task/internal/repository"

	"github.com/gin-gonic/gin"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context) (models.Wallet, error)
	FundTranswer(ctx context.Context, fromID, toID string, amount float32) error
	GetWalletBalance(ctx context.Context, id string) (float32, error)
	GetWalletHistory(ctx context.Context, id string) ([]models.Transaction, error)
	IsWalletExists(ctx context.Context, id string) (bool, error)
}

type Handler struct {
	repo WalletRepository
}

func NewHandler(repo WalletRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.New()

	api := r.Group("/api/v1")
	{
		api.POST("/wallet", h.CreateWallet)
		api.POST("/wallet/:walletId/send", h.FundTranswer)
		api.GET("/wallet/:walletId", h.GetWalletStatus)
		api.GET("/wallet/:walletId/history", h.GetHistory)
	}

	return r
}

func (h *Handler) CreateWallet(c *gin.Context) {
	wallet, err := h.repo.CreateWallet(c.Request.Context())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

type fundTranswerRequest struct {
	To     string  `json:"to"`
	Amount float32 `json:"amount"`
}

func (h *Handler) FundTranswer(c *gin.Context) {
	from := c.Param("walletId")

	var req fundTranswerRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json body"})
		return
	}

	if from == req.To {
		c.JSON(http.StatusBadRequest, gin.H{"message": "destination and from musn't be equal"})
		return
	}
	if req.Amount <= 0.0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "amount musn't be less or equal zero"})
		return
	}

	balance, err := h.repo.GetWalletBalance(c.Request.Context(), from)
	if errors.Is(err, repository.ErrWalletNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if balance < req.Amount {
		c.Status(http.StatusBadRequest)
		return
	}

	err = h.repo.FundTranswer(c.Request.Context(), from, req.To, req.Amount)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) GetWalletStatus(c *gin.Context) {
	id := c.Param("walletId")
	balance, err := h.repo.GetWalletBalance(c.Request.Context(), id)
	if errors.Is(err, repository.ErrWalletNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	wallet := models.Wallet{
		Id:      id,
		Balance: balance,
	}

	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) GetHistory(c *gin.Context) {
	id := c.Param("walletId")

	exist, err := h.repo.IsWalletExists(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else if !exist {
		c.Status(http.StatusNotFound)
		return
	}

	history, err := h.repo.GetWalletHistory(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, history)
}
