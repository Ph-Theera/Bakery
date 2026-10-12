package rest

import (
	catalogin "bakery/internal/ports/port_in/catalog"
	ordersin "bakery/internal/ports/port_in/orders"
	storein "bakery/internal/ports/port_in/store"
	usersin "bakery/internal/ports/port_in/users"
)

type Usecases struct {
	ListProducts      catalogin.ListProductsUsecase
	GetProduct        catalogin.GetProductUsecase
	CreateProduct     catalogin.CreateProductUsecase
	UpdateProduct     catalogin.UpdateProductUsecase
	DeleteProduct     catalogin.DeleteProductUsecase
	ListCategories    catalogin.ListCategoriesUsecase
	QuoteCheckout     ordersin.QuoteCheckoutUsecase
	CreateOrder       ordersin.CreateOrderUsecase
	GetOrder          ordersin.GetOrderUsecase
	ListOrders        ordersin.ListOrdersUsecase
	ListMyOrders      ordersin.ListMyOrdersUsecase
	CheckOrderAccess  ordersin.CheckOrderAccessUsecase
	UpdateOrderStatus ordersin.UpdateOrderStatusUsecase
	UpdateOrderNote   ordersin.UpdateOrderNoteUsecase
	GetDashboard      ordersin.GetDashboardUsecase
	GetPaymentOrder   ordersin.GetPaymentOrderUsecase
	UploadPaymentSlip ordersin.UploadPaymentSlipUsecase
	SubmitPayment     ordersin.SubmitPaymentUsecase
	GetPaymentSlip    ordersin.GetPaymentSlipUsecase
	ApprovePayment    ordersin.ApprovePaymentUsecase
	RejectPayment     ordersin.RejectPaymentUsecase
	GetStore          storein.GetStoreUsecase
	UpdateStore       storein.UpdateStoreUsecase
	ListCustomers     usersin.ListCustomersUsecase
	LinkLine          usersin.LinkLineUsecase
}
