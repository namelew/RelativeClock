package minheap

import (
	"errors"

	"golang.org/x/exp/constraints"
)

type Ordened[T constraints.Integer] interface {
	Value() T
}

type MinHeap[K constraints.Integer] struct {
	data []Ordened[K]
	size int
}

func New[K constraints.Integer]() *MinHeap[K] {
	return &MinHeap[K]{
		data: make([]Ordened[K], 0),
		size: 0,
	}
}

func (h *MinHeap[K]) Insert(x Ordened[K]) {
	h.data = append(h.data, x)
	h.size++
	h.bubbleUp(h.size - 1)
}

func (h *MinHeap[K]) bubbleUp(i int) {
	if i == 0 {
		return
	}
	p := (i - 1) / 2
	if h.data[i].Value() < (h.data[p]).Value() {
		h.data[i], h.data[p] = h.data[p], h.data[i]
		h.bubbleUp(p)
	}
}

func (h *MinHeap[K]) ExtractMin() (Ordened[K], error) {
	if h.IsEmpty() {
		return nil, errors.New("heap is empty")
	}
	min := h.data[0]
	h.data[0] = h.data[h.size-1]
	h.data = h.data[:h.size-1]
	h.size--
	h.bubbleDown(0)
	return min, nil
}

func (h *MinHeap[K]) bubbleDown(i int) {
	l := 2*i + 1
	r := 2*i + 2

	smallest := i
	if l < h.size && h.data[l].Value() < h.data[smallest].Value() {
		smallest = l
	}
	if r < h.size && h.data[r].Value() < h.data[smallest].Value() {
		smallest = r
	}

	if smallest != i {
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		h.bubbleDown(smallest)
	}
}

func (h *MinHeap[K]) Peek() (Ordened[K], error) {
	if h.IsEmpty() {
		return nil, errors.New("heap is empty")
	}
	return h.data[0], nil
}

func (h *MinHeap[K]) IsEmpty() bool {
	return h.size == 0
}
