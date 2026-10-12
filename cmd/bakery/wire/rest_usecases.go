package wire

import rest "bakery/internal/adapters/inbound/rest/v1"

func (a *Application) NewRestUsecases() *rest.Usecases {
	return &rest.Usecases{
		ListProducts:      a.listProductsUc,
		GetProduct:        a.getProductUc,
		CreateProduct:     a.createProductUc,
		UpdateProduct:     a.updateProductUc,
		DeleteProduct:     a.deleteProductUc,
		ListCategories:    a.listCategoriesUc,
		QuoteCheckout:     a.quoteCheckoutUc,
		CreateOrder:       a.createOrderUc,
		GetOrder:          a.getOrderUc,
		ListOrders:        a.listOrdersUc,
		ListMyOrders:      a.listMyOrdersUc,
		CheckOrderAccess:  a.checkOrderAccessUc,
		UpdateOrderStatus: a.updateOrderStatusUc,
		UpdateOrderNote:   a.updateOrderNoteUc,
		GetDashboard:      a.getDashboardUc,
		GetPaymentOrder:   a.getPaymentOrderUc,
		UploadPaymentSlip: a.uploadPaymentSlipUc,
		SubmitPayment:     a.submitPaymentUc,
		GetPaymentSlip:    a.getPaymentSlipUc,
		ApprovePayment:    a.approvePaymentUc,
		RejectPayment:     a.rejectPaymentUc,
		GetStore:          a.getStoreUc,
		UpdateStore:       a.updateStoreUc,
		ListCustomers:     a.listCustomersUc,
		LinkLine:          a.linkLineUc,
	}
}
