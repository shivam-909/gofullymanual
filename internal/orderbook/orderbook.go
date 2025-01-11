package orderbook

import "math/rand/v2"

type OrderSide int

const (
	OrderSideBuy  = 1
	OrderSideSell = 2
	MaxPrice      = 10000
	MinPrice      = 9000
)

var counter int = 1
var removals int = 1

type Order struct {
	Id    int
	Side  OrderSide
	Price int
	Qty   int
}

type OrderBook interface {
	Insert(order Order) error
	Remove(id int) error
}

type OrderBookNode struct {
	Order Order
	Left  *OrderBookNode
	Right *OrderBookNode
}

func randomBool() bool {
	n := rand.Int()
	return n%2 == 0
}

func randomBoolDistribution(truePercentage int) bool {
	n := rand.IntN(100)
	return n < truePercentage
}

func randomSide() OrderSide {
	if randomBool() {
		return OrderSideBuy
	}
	return OrderSideSell
}

func GenerateOrder() Order {
	side := randomSide()
	price := rand.IntN(MaxPrice-MinPrice) + MinPrice
	qty := rand.IntN(10) + 1
	id := counter
	counter++
	return Order{id, side, price, qty}
}

func Act(ob OrderBook) {
	if randomBoolDistribution(50) {
		ord := GenerateOrder()
		ob.Insert(ord)
	} else {
		if removals > counter {
			return
		}
		ob.Remove(removals)
		removals++
	}
}
