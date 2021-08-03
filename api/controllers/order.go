package controllers

import (
	"net/http"
	"strconv"
	"travel/api/responses"
	"travel/api/services"
	"travel/constants"
	"travel/infrastructure"
	"travel/models"
	"travel/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderController -> data type
type OrderController struct {
	logger       infrastructure.Logger
	orderService services.OrderService
}

// NewOrderController -> creates new user controller
func NewOrderController(logger infrastructure.Logger, orderService services.OrderService, firebaseService services.FirebaseService) OrderController {
	return OrderController{
		logger:       logger,
		orderService: orderService,
	}
}

// GetAllOrders
func (cc OrderController) GetAllOrders(c *gin.Context) {
	pagination := utils.BuildPagination(c)
	searchParams := models.OrderSearchParams{
		Keyword: c.Query("keyword"),
	}
	orders, count, err := cc.orderService.GetAllOrders(searchParams, pagination)
	if err != nil {
		cc.logger.Zap.Error("Failed to get orders::", err)
		responses.ErrorJSON(c, http.StatusBadRequest, "Failed to get orders")
		return
	}
	responses.JSONCount(c, http.StatusOK, orders, int(count))
}

// CreateOrder -> create a new order controller
func (o OrderController) CreateOrder(c *gin.Context) {
	var order models.Order
	trx := c.MustGet(constants.DBTransaction).(*gorm.DB)
	if err := c.ShouldBindJSON(&order); err != nil {
		o.logger.Zap.Error("Failed to bind order json data::", err.Error())
		responses.ErrorJSON(c, http.StatusBadRequest, "Failed to parse json data")
		return
	}
	// Check if items are available in Order object as order_items:
	if len(order.OrderItem) > 0 {
		// calculate the total
		var total_amount float64 = 0
		var sub_total float64 = 0
		for _, item := range order.OrderItem {
			sub_total = item.Price * float64(item.Quantity)
			total_amount += sub_total
		}
		order.TotalAmount = total_amount

		orderObj, err := o.orderService.CreateOrder(order)
		if err != nil {
			o.logger.Zap.Error("Failed to create order ::", err.Error())
			responses.ErrorJSON(c, http.StatusBadRequest, "Failed to save order")
			return
		}

		for _, orderItem := range order.OrderItem {
			orderItem.OrderID = int(orderObj.ID)
			err = o.orderService.WithTrx(trx).CreateOrderItem(orderItem)
			if err != nil {
				o.logger.Zap.Error("Error [CreateOrderItemService] ::", err.Error())
				responses.ErrorJSON(c, http.StatusBadRequest, "Failed to save order")
				return
			}
		}
	} else {
		o.logger.Zap.Error("Error [No order Item.] ::")
		responses.ErrorJSON(c, http.StatusBadRequest, "Failed to save order becasue no order_items is sent")
		return
	}

	responses.SuccessJSON(c, http.StatusCreated, "Order created successfully")
}

//GetOrderByID -> returns a order by ID
func (o OrderController) GetOrderByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		o.logger.Zap.Error("Error retriving id param:", err)
		responses.ErrorJSON(c, http.StatusBadRequest, "Failed to retrieve id param")
		return
	}
	orderObj := models.Order{}
	orderObj.ID = uint(id)
	order, err := o.orderService.GetOrderByID(orderObj)
	if err != nil {
		o.logger.Zap.Error("Can not find order:", err)
		responses.ErrorJSON(c, http.StatusBadRequest, "Can not find order")
		return
	}
	responses.JSON(c, http.StatusOK, order)
}