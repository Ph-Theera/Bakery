package wire

import (
	"bakery/cmd/bakery/wire/dep"
	"bakery/infrastructure/config"
	"bakery/internal/adapters/outbound/db/catalog"
	inventorydb "bakery/internal/adapters/outbound/db/inventory"
	notificationsdb "bakery/internal/adapters/outbound/db/notifications"
	ordersdb "bakery/internal/adapters/outbound/db/orders"
	paymentsdb "bakery/internal/adapters/outbound/db/payments"
	storedb "bakery/internal/adapters/outbound/db/store"
	"bakery/internal/adapters/outbound/db/transaction"
	usersdb "bakery/internal/adapters/outbound/db/users"
	"bakery/internal/adapters/outbound/line"
	catalogin "bakery/internal/ports/port_in/catalog"
	notificationsin "bakery/internal/ports/port_in/notifications"
	ordersin "bakery/internal/ports/port_in/orders"
	storein "bakery/internal/ports/port_in/store"
	usersin "bakery/internal/ports/port_in/users"
	"bakery/internal/services/checkout"
	compositeuc "bakery/internal/usecases/composite"
	ordersuc "bakery/internal/usecases/orders"
	paymentsuc "bakery/internal/usecases/payments"
	productuc "bakery/internal/usecases/product"
	storeuc "bakery/internal/usecases/store"
	usersuc "bakery/internal/usecases/users"
	"net/http"
	"time"
)

type Application struct {
	listProductsUc          catalogin.ListProductsUsecase
	getProductUc            catalogin.GetProductUsecase
	createProductUc         catalogin.CreateProductUsecase
	updateProductUc         catalogin.UpdateProductUsecase
	deleteProductUc         catalogin.DeleteProductUsecase
	listCategoriesUc        catalogin.ListCategoriesUsecase
	quoteCheckoutUc         ordersin.QuoteCheckoutUsecase
	createOrderUc           ordersin.CreateOrderUsecase
	getOrderUc              ordersin.GetOrderUsecase
	listOrdersUc            ordersin.ListOrdersUsecase
	listMyOrdersUc          ordersin.ListMyOrdersUsecase
	checkOrderAccessUc      ordersin.CheckOrderAccessUsecase
	updateOrderStatusUc     ordersin.UpdateOrderStatusUsecase
	updateOrderNoteUc       ordersin.UpdateOrderNoteUsecase
	getDashboardUc          ordersin.GetDashboardUsecase
	getPaymentOrderUc       ordersin.GetPaymentOrderUsecase
	uploadPaymentSlipUc     ordersin.UploadPaymentSlipUsecase
	submitPaymentUc         ordersin.SubmitPaymentUsecase
	getPaymentSlipUc        ordersin.GetPaymentSlipUsecase
	approvePaymentUc        ordersin.ApprovePaymentUsecase
	rejectPaymentUc         ordersin.RejectPaymentUsecase
	getStoreUc              storein.GetStoreUsecase
	updateStoreUc           storein.UpdateStoreUsecase
	listCustomersUc         usersin.ListCustomersUsecase
	linkLineUc              usersin.LinkLineUsecase
	dispatchNotificationsUc notificationsin.DispatchNotificationsUsecase
}

func NewApplication(deps dep.Dependencies, c config.Config) *Application {
	products := catalog.NewProductRepository(deps.Db, c.ImageBase)
	orders := ordersdb.NewOrderRepository(deps.Db)
	payments := paymentsdb.NewPaymentRepository(deps.Db)
	inventory := inventorydb.NewInventoryRepository(deps.Db)
	users := usersdb.NewUserRepository(deps.Db)
	store := storedb.NewStoreRepository(deps.Db)
	outbox := notificationsdb.NewOutboxRepository(deps.Db)
	exec := transaction.NewExecWithTx(deps.Db)
	client := &http.Client{Timeout: 5 * time.Second}
	identity := line.Identity{Client: client, ChannelID: c.LineLoginID}
	pricing := checkout.Pricing{Products: products, Inventory: inventory, Store: store}
	review := compositeuc.PaymentReviewArgs{ExecWithTx: exec, OrderRepository: orders, PaymentRepository: payments, InventoryRepository: inventory, CatalogRepository: products, Outbox: outbox}
	return &Application{
		listProductsUc:          productuc.NewListProductsUsecase(productuc.ListProductsArgs{CatalogRepository: products}),
		getProductUc:            productuc.NewGetProductUsecase(productuc.GetProductArgs{CatalogRepository: products}),
		createProductUc:         productuc.NewCreateProductUsecase(productuc.CreateProductArgs{CatalogRepository: products}),
		updateProductUc:         productuc.NewUpdateProductUsecase(productuc.UpdateProductArgs{ExecWithTx: exec, CatalogRepository: products, InventoryRepository: inventory}),
		deleteProductUc:         productuc.NewDeleteProductUsecase(productuc.DeleteProductArgs{CatalogRepository: products}),
		listCategoriesUc:        productuc.NewListCategoriesUsecase(productuc.ListCategoriesArgs{CatalogRepository: products}),
		quoteCheckoutUc:         compositeuc.NewQuoteCheckoutUsecase(compositeuc.QuoteCheckoutArgs{Pricing: pricing}),
		createOrderUc:           compositeuc.NewCreateOrderUsecase(compositeuc.CreateOrderArgs{ExecWithTx: exec, OrderRepository: orders, UserRepository: users, PaymentRepository: payments, InventoryRepository: inventory, Pricing: pricing}),
		getOrderUc:              ordersuc.NewGetOrderUsecase(ordersuc.GetOrderArgs{OrderRepository: orders}),
		listOrdersUc:            ordersuc.NewListOrdersUsecase(ordersuc.ListOrdersArgs{OrderRepository: orders}),
		listMyOrdersUc:          ordersuc.NewListMyOrdersUsecase(ordersuc.ListMyOrdersArgs{OrderRepository: orders}),
		checkOrderAccessUc:      ordersuc.NewCheckOrderAccessUsecase(ordersuc.CheckOrderAccessArgs{OrderRepository: orders}),
		updateOrderStatusUc:     compositeuc.NewUpdateOrderStatusUsecase(compositeuc.UpdateOrderStatusArgs{ExecWithTx: exec, OrderRepository: orders, PaymentRepository: payments, InventoryRepository: inventory}),
		updateOrderNoteUc:       ordersuc.NewUpdateOrderNoteUsecase(ordersuc.UpdateOrderNoteArgs{OrderRepository: orders}),
		getDashboardUc:          ordersuc.NewGetDashboardUsecase(ordersuc.GetDashboardArgs{OrderRepository: orders}),
		getPaymentOrderUc:       paymentsuc.NewGetPaymentOrderUsecase(paymentsuc.GetPaymentOrderArgs{PaymentRepository: payments}),
		uploadPaymentSlipUc:     compositeuc.NewUploadPaymentSlipUsecase(compositeuc.UploadPaymentSlipArgs{ExecWithTx: exec, OrderRepository: orders, PaymentRepository: payments}),
		submitPaymentUc:         compositeuc.NewSubmitPaymentUsecase(compositeuc.SubmitPaymentArgs{ExecWithTx: exec, OrderRepository: orders, PaymentRepository: payments, InventoryRepository: inventory}),
		getPaymentSlipUc:        paymentsuc.NewGetPaymentSlipUsecase(paymentsuc.GetPaymentSlipArgs{PaymentRepository: payments}),
		approvePaymentUc:        compositeuc.NewApprovePaymentUsecase(review),
		rejectPaymentUc:         compositeuc.NewRejectPaymentUsecase(review),
		getStoreUc:              storeuc.NewGetStoreUsecase(storeuc.GetStoreArgs{StoreRepository: store}),
		updateStoreUc:           storeuc.NewUpdateStoreUsecase(storeuc.UpdateStoreArgs{StoreRepository: store}),
		listCustomersUc:         usersuc.NewListCustomersUsecase(usersuc.ListCustomersArgs{UserRepository: users}),
		linkLineUc:              usersuc.NewLinkLineUsecase(usersuc.LinkLineArgs{UserRepository: users, Identity: identity}),
		dispatchNotificationsUc: compositeuc.NewDispatchNotificationsUsecase(compositeuc.DispatchNotificationsArgs{Outbox: outbox, Sender: line.Sender{Client: client, Token: c.LineToken, Recipient: c.LineRecipient}}),
	}
}
