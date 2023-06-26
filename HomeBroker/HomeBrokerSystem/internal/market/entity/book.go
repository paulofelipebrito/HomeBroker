package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Order         []*Order
	Transactions   []*Transaction
	OrdersChan    chan *Order
	OrdersChanOut chan *Order
	Wg            *sync.WaitGroup
}

func NewBook(ordersChan chan *Order, ordersChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Order: []*Order{},
		Transactions: []*Transaction{},
		OrdersChan: ordersChan,
		OrdersChanOut: ordersChanOut,
		Wg: wg,
	}
}

// func (book *Book) Trade() {
// 	buyOrders := NewOrderQueue()
// 	sellOrders := NewOrderQueue()

// 	heap.Init(buyOrders)
// 	heap.Init(sellOrders)

// 	for order := range book.OrdersChan {
// 		if order.OrderType == "BUY"{
// 			buyOrders.Push(order)
// 			if sellOrders.Len() > 0 && sellOrders.Orders[0].Price <= order.Price {
// 				sellOrder := sellOrders.Pop().(*Order)

// 				if sellOrder.PendingShares > 0 {
// 					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
// 					book.AddTransaction(transaction, book.Wg)
// 					sellOrder.Transations = append(sellOrder.Transations, transaction)
// 					order.Transations = append(order.Transations, transaction)
// 					book.OrdersChanOut <- sellOrder
// 					book.OrdersChanOut <- order

// 					if sellOrder.PendingShares > 0 {
// 						sellOrders.Push(sellOrder)
// 					}
// 				} 
// 			} 
// 		} else {
// 			sellOrders.Push(order)

// 			if buyOrders.Len() > 0 && buyOrders.Orders[0].Price <= order.Price {
// 				buyOrder := buyOrders.Pop().(*Order)

// 				if buyOrder.PendingShares > 0 {
// 					transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
// 					book.AddTransaction(transaction, book.Wg)
// 					buyOrder.Transations = append(buyOrder.Transations, transaction)
// 					order.Transations = append(order.Transations, transaction)
// 					book.OrdersChanOut <- buyOrder
// 					book.OrdersChanOut <- order

// 					if buyOrder.PendingShares > 0 {
// 						buyOrders.Push(buyOrder)
// 					}
// 				} 
// 			} 
// 		}
// 	}
// }

func (book *Book) Trade() {
	// buyOrders := NewOrderQueue()
	// sellOrders := NewOrderQueue()
	buyOrders := make(map[string]*OrderQueue)
	sellOrders := make(map[string]*OrderQueue)

	// heap.Init(buyOrders)
	// heap.Init(sellOrders)

	for order := range book.OrdersChan {
		asset := order.Asset.ID

		if buyOrders[asset] == nil {
			buyOrders[asset] = NewOrderQueue()
			heap.Init(buyOrders[asset])
		}

		if sellOrders[asset] == nil {
			sellOrders[asset] = NewOrderQueue()
			heap.Init(sellOrders[asset])
		}

		if order.OrderType == "BUY" {
			processOrder(order, sellOrders, buyOrders, book, asset)
		} else if order.OrderType == "SELL"{
			processOrder(order, buyOrders, sellOrders, book, asset)
		}
	}
}

func processOrder(order *Order, orderToCheckShares map[string]*OrderQueue, orderToAddInQueue map[string]*OrderQueue, book *Book, asset string) {
	orderToAddInQueue[asset].Push(order)

	if orderToCheckShares[asset].Len() > 0 && orderToCheckShares[asset].Orders[0].Price <= order.Price {
		oppositeOrder := orderToCheckShares[asset].Pop().(*Order)

		if oppositeOrder.PendingShares > 0 {
			transaction := NewTransaction(oppositeOrder, order, order.Shares, oppositeOrder.Price)
			book.AddTransaction(transaction, book.Wg)
			oppositeOrder.Transactions = append(oppositeOrder.Transactions, transaction)
			order.Transactions = append(order.Transactions, transaction)
			book.OrdersChanOut <- oppositeOrder
			book.OrdersChanOut <- order

			if oppositeOrder.PendingShares > 0 {
				orderToCheckShares[asset].Push(oppositeOrder)
			}
		}
	}
}

func (book *Book) AddTransaction (transaction *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	minShares := sellingShares
	if buyingShares < minShares {
		minShares = buyingShares
	}

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	transaction.AddSellOrderPendingShares(-minShares)

	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.BuyingOrder.Asset.ID, minShares)
	transaction.AddBuyOrderPendingShares(-minShares)

	transaction.CalculateTotal(transaction.Shares, transaction.BuyingOrder.Price)
	transaction.CloseBuyOrder()
	transaction.CloseSellOrder()
	book.Transactions = append(book.Transactions, transaction)
}