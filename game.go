package main

// Cell represents state of single active cell
type Cell struct {
	Row, Col int
}

// Board represents the current state of the game board for Conway's Game Of Life.
// Conway's Game of Life takes place on an infinite two-dimensional board of square cells. Each cell is either dead or alive, and at each tick, the following rules apply:
//
// Any live cell with less than two live neighbours dies.
// Any live cell with two or three live neighbours remains living.
// Any live cell with more than three live neighbours dies.
// Any dead cell with exactly three live neighbours becomes a live cell.
// A cell neighbours another cell if it is horizontally, vertically, or diagonally adjacent.
type Board struct {
	ActiveCells map[Cell]struct{}
}

// NewBoard returns a non initialized instance of game board
func NewBoard(activeCells []Cell) *Board {
	b := &Board{ActiveCells: make(map[Cell]struct{})}
	for _, c := range activeCells {
		b.ActiveCells[c] = struct{}{}
	}
	return b
}

// Step computes the next state of the board based on it's the current state. It updates the board's state to the computed state.
// We use BFS to explore the active cells and their neighbors and apply the game's rule to determine their next state.
func (b *Board) Step() {
	queued := map[Cell]struct{}{}
	newActiveCells := make(map[Cell]struct{})
	for c := range b.ActiveCells {
		if _, ok := queued[c]; !ok {
			b.check(c, newActiveCells, queued)
		}
	}
	b.ActiveCells = newActiveCells
}

func (b *Board) check(c Cell, newActiveCells, queued map[Cell]struct{}) {
	shouldLive := func(c Cell) bool {
		neighs := []Cell{
			{c.Row - 1, c.Col - 1}, {c.Row - 1, c.Col}, {c.Row - 1, c.Col + 1},
			{c.Row, c.Col - 1}, {c.Row, c.Col + 1},
			{c.Row + 1, c.Col - 1}, {c.Row + 1, c.Col}, {c.Row + 1, c.Col + 1},
		}
		activeNeighs := 0
		for _, neigh := range neighs {
			if _, ok := b.ActiveCells[neigh]; ok {
				activeNeighs++
			}
		}
		if _, ok := b.ActiveCells[c]; ok && activeNeighs >= 2 && activeNeighs <= 3 {
			return true
		}
		if _, ok := b.ActiveCells[c]; !ok && activeNeighs == 3 {
			return true
		}
		return false
	}

	q := []Cell{c}
	queued[c] = struct{}{}
	for len(q) > 0 {
		c := q[0]
		q = q[1:]

		if shouldLive(c) {
			newActiveCells[c] = struct{}{}
		}
		if _, ok := b.ActiveCells[c]; !ok {
			// we only need to consider current active cells and their neighbors
			continue
		}
		neighs := []Cell{
			{c.Row - 1, c.Col - 1}, {c.Row - 1, c.Col}, {c.Row - 1, c.Col + 1},
			{c.Row, c.Col - 1}, {c.Row, c.Col + 1},
			{c.Row + 1, c.Col - 1}, {c.Row + 1, c.Col}, {c.Row + 1, c.Col + 1},
		}
		for _, neigh := range neighs {
			if _, ok := queued[neigh]; !ok {
				q = append(q, neigh)
				queued[neigh] = struct{}{}
			}
		}
	}
}
