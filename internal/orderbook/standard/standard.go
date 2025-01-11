package standardbook

import (
	"errors"

	"github.com/shivam-909/gofullymanual/internal/orderbook"
)

type standardbook struct {
	root *orderbook.OrderBookNode
}

func New() orderbook.OrderBook {
	return &standardbook{}
}

// Insert places an Order into the BST by Price.
// - If the tree is empty, newNode becomes the root.
// - Otherwise, we walk left/right until we find a spot.
// - Duplicates (same price) go to the right.
func (b *standardbook) Insert(o orderbook.Order) error {
	newNode := &orderbook.OrderBookNode{
		Order: o,
	}

	if b.root == nil {
		b.root = newNode
		return nil
	}

	curr := b.root
	for {
		if o.Price < curr.Order.Price {
			if curr.Left == nil {
				curr.Left = newNode
				return nil
			}
			curr = curr.Left
		} else {
			if curr.Right == nil {
				curr.Right = newNode
				return nil
			}
			curr = curr.Right
		}
	}
}

// Remove locates a node by its Price and removes it from the BST.
func (b *standardbook) Remove(price int) error {
	parent, node, wentLeft := b.findNodeByPrice(price)
	if node == nil {
		return errors.New("order not found")
	}

	var replacement *orderbook.OrderBookNode

	switch {
	case node.Left == nil:
		replacement = node.Right
	case node.Right == nil:
		replacement = node.Left
	default:
		succParent, successor := b.findSuccessor(node.Right)

		if succParent != nil && succParent != node {
			succParent.Left = successor.Right
			successor.Right = node.Right
		}
		successor.Left = node.Left
		replacement = successor
	}

	if parent == nil {
		b.root = replacement
	} else if wentLeft {
		parent.Left = replacement
	} else {
		parent.Right = replacement
	}

	return nil
}

// findNodeByPrice walks the tree to locate the node with matching price,
// returning:
//   - parent of the found node (or nil if node is the root)
//   - the node with the given price (or nil if not found)
//   - a bool indicating if the node was a left child of its parent
func (b *standardbook) findNodeByPrice(price int) (parent, found *orderbook.OrderBookNode, isLeft bool) {
	var (
		curr   = b.root
		par    *orderbook.OrderBookNode
		leftCh bool
	)

	for curr != nil {
		if price == curr.Order.Price {
			return par, curr, leftCh
		}
		par = curr
		if price < curr.Order.Price {
			curr = curr.Left
			leftCh = true
		} else {
			curr = curr.Right
			leftCh = false
		}
	}
	return nil, nil, false
}

// findSuccessor finds the leftmost node of the given subtree 'node'
// and returns: (parentOfSuccessor, successorNode).
func (b *standardbook) findSuccessor(node *orderbook.OrderBookNode) (parent, successor *orderbook.OrderBookNode) {
	if node == nil {
		return nil, nil
	}

	var (
		par  *orderbook.OrderBookNode
		curr = node
	)

	for curr.Left != nil {
		par = curr
		curr = curr.Left
	}

	return par, curr
}
