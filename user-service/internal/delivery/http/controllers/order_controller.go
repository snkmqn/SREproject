package controllers

import (
	"user-service/internal/core/models"
	"user-service/internal/usecases/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type OrderController struct {
	orderService services.OrderService
}

func NewOrderController(orderService services.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var orderRequest struct {
		UserID string `json:"user_id" binding:"required"`
		Status string `json:"status" binding:"required"`
		Items  []struct {
			ProductID    string  `json:"product_id" binding:"required"`
			Quantity     int     `json:"quantity" binding:"required"`
			PricePerUnit float64 `json:"price_per_unit" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if orderRequest.Status != "pending" && orderRequest.Status != "completed" && orderRequest.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Allowed values are 'pending', 'completed', 'cancelled'."})
		return
	}
	for _, item := range orderRequest.Items {
		if item.Quantity < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity cannot be negative"})
			return
		}
		if item.PricePerUnit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Price per unit cannot be negative"})
			return
		}
	}

	items := []models.OrderItem{}
	for _, item := range orderRequest.Items {
		items = append(items, models.OrderItem{
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			PricePerUnit: item.PricePerUnit,
		})
	}

	userID := c.GetString("user_id")

	order, err := ctrl.orderService.CreateOrder(c, userID, items, orderRequest.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order": order})
}

func (ctrl *OrderController) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	order, err := ctrl.orderService.GetOrderByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (ctrl *OrderController) UpdateOrder(c *gin.Context) {
	id := c.Param("id")

	var orderUpdate struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&orderUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	err := ctrl.orderService.UpdateOrder(c, id, orderUpdate.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}


func (ctrl *OrderController) GetOrdersByUserID(c *gin.Context) {
	userID := c.Param("id")
	log.Println("Fetching order with ID:", userID)

	orders, err := ctrl.orderService.GetOrdersByUserID(c, userID)
	if err != nil {
		log.Println("Error fetching order:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No orders found for this user"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
