package markdown

/*
Elements are not allocated one at a time, but in rows of
elemHeap.RowSize elements. After N elements have been
requested, a row is exhausted, and the next one will
be allocated. Previously allocated rows are tracked in
elemHeap.rows.

Pos() and setPos() methods allow to query and reset the
current position (row, and position within the row), which
allows reusing elements. It must be made sure, that previous
users of such storage don't access it anymore once setPos has
been called.
*/

type elemHeap struct {
	rows [][]element
	heapPos
	rowSize int
}

type heapPos struct {
	iRow int
	row  []element
}

func (h *elemHeap) nextRow() []element {
	h.iRow++
	if h.iRow == len(h.rows) {
		h.rows = append(h.rows, make([]element, h.rowSize))
	}
	h.row = h.rows[h.iRow]
	return h.row
}

func (h *elemHeap) init(size int) {
	h.rowSize = size
	h.rows = [][]element{make([]element, size)}
	h.row = h.rows[h.iRow]
}

func (h *elemHeap) Pos() heapPos {
	return h.heapPos
}

func (h *elemHeap) setPos(i heapPos) {
	h.heapPos = i
}
