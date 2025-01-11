package manualbook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shivam-909/gofullymanual/alloc"
	"github.com/shivam-909/gofullymanual/internal/orderbook"
)

// manualbook implements orderbook.OrderBook using a BST
// and manual allocations (via alloc).
type manualbook struct {
	tree *orderbook.OrderBookNode
}

// New creates a new manualbook instance.
func New() orderbook.OrderBook {
	return &manualbook{}
}

// newNode allocates a new OrderBookNode for the given Order
// from your manual allocator.
func newNode(o orderbook.Order) *orderbook.OrderBookNode {
	node := alloc.Allocate[orderbook.OrderBookNode](1)
	if node == nil {
		return nil
	}
	node.Order = o
	node.Left = nil
	node.Right = nil
	return node
}

// Insert adds a new order into the BST keyed by Order.Id.
// Duplicates (same ID) go to the right.
func (b *manualbook) Insert(o orderbook.Order) error {
	nn := newNode(o)
	if nn == nil {
		return errors.New("allocation failed")
	}

	if b.tree == nil {
		b.tree = nn
		return nil
	}

	curr := b.tree
	for {
		if o.Id < curr.Order.Id {
			if curr.Left == nil {
				curr.Left = nn
				return nil
			}
			curr = curr.Left
		} else {
			if curr.Right == nil {
				curr.Right = nn
				return nil
			}
			curr = curr.Right
		}
	}
}

// Remove locates a node by its Order.Id and removes it from the BST.
// If a node to remove has two children, we use the in-order successor
// (the leftmost node in its right subtree). This logic is carefully
// done to avoid creating cycles.
func (b *manualbook) Remove(id int) error {
	parent, node, isLeft := b.findNodeById(id)
	if node == nil {
		return errors.New("order not found")
	}

	var replacement *orderbook.OrderBookNode

	switch {
	// Case 1: node is a leaf
	case node.Left == nil && node.Right == nil:
		replacement = nil

	// Case 2: node has only a right subtree
	case node.Left == nil:
		replacement = node.Right

	// Case 3: node has only a left subtree
	case node.Right == nil:
		replacement = node.Left

	// Case 4: node has both left & right subtrees => find successor
	default:
		succParent, successor := b.findSuccessor(node.Right)

		// If successor has a parent that isn't this node,
		// then we detach successor from that parent's left
		// and reattach successor's right subtree there.
		if succParent != nil && succParent != node {
			succParent.Left = successor.Right
			successor.Right = node.Right
		}

		// The successor always takes the left subtree
		successor.Left = node.Left
		replacement = successor
	}

	// Now link 'replacement' into the tree
	if parent == nil {
		b.tree = replacement // node was root
	} else if isLeft {
		parent.Left = replacement
	} else {
		parent.Right = replacement
	}

	// Free the removed node
	alloc.Free(node, 1)
	return nil
}

// findNodeById walks the BST to find the node whose Order.Id == id.
// Returns (parent, node, isLeftChild).
func (b *manualbook) findNodeById(id int) (*orderbook.OrderBookNode, *orderbook.OrderBookNode, bool) {
	var (
		parent  *orderbook.OrderBookNode
		current = b.tree
		isLeft  bool
	)

	for current != nil {
		if id == current.Order.Id {
			return parent, current, isLeft
		}
		parent = current
		if id < current.Order.Id {
			current = current.Left
			isLeft = true
		} else {
			current = current.Right
			isLeft = false
		}
	}
	return nil, nil, false
}

// findSuccessor returns (parent, successor) for the leftmost node in 'root'.
// Called by Remove to find the in-order successor in node.Right.
func (b *manualbook) findSuccessor(root *orderbook.OrderBookNode) (*orderbook.OrderBookNode, *orderbook.OrderBookNode) {
	if root == nil {
		return nil, nil
	}
	var (
		parent *orderbook.OrderBookNode
		curr   = root
	)
	for curr.Left != nil {
		parent = curr
		curr = curr.Left
	}
	return parent, curr
}

// Print traverses all orders and prints them, grouping by side.
func (b *manualbook) Print() {
	var buyOrders, sellOrders []orderbook.Order

	var collect func(node *orderbook.OrderBookNode)
	collect = func(node *orderbook.OrderBookNode) {
		if node == nil {
			return
		}
		collect(node.Left)
		if node.Order.Side == orderbook.OrderSideBuy {
			buyOrders = append(buyOrders, node.Order)
		} else {
			sellOrders = append(sellOrders, node.Order)
		}
		collect(node.Right)
	}
	collect(b.tree)

	fmt.Println("\nOrder Book")
	fmt.Println(strings.Repeat("-", 40))

	fmt.Println("Sells:")
	for i := len(sellOrders) - 1; i >= 0; i-- {
		o := sellOrders[i]
		fmt.Printf("Price: %d, Quantity: %d, ID: %d\n", o.Price, o.Qty, o.Id)
	}

	fmt.Println(strings.Repeat("-", 40))

	fmt.Println("Buys:")
	for i := len(buyOrders) - 1; i >= 0; i-- {
		o := buyOrders[i]
		fmt.Printf("Price: %d, Quantity: %d, ID: %d\n", o.Price, o.Qty, o.Id)
	}
	fmt.Println(strings.Repeat("-", 40))
}
