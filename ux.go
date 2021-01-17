package main

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"syscall/js"
	"text/template"
	"time"
)

var boardTmplt = `
<style>
.tableLayout {
	margin-left: auto;
	margin-right: auto;
	table-layout: fixed;
}
.tableHeader {
	background-color: tomato;
	  color: black;
	  padding: 10px;
	  text-align: center;
}
.newCell {
	background-color: darkseagreen;
	color: black;
	text-align: center;
	width: 30px;
	height: 30px;
}
.liveCell {
	background-color: lightseagreen;
	color: black;
	text-align: center;
	width: 30px;
	height: 30px;
}
.deadCell {
	background-color: grey;
	color: black;
	text-align: center;
	width: 30px;
	height: 30px;
}
</style>
<table class=tableLayout>
	<th vlass=tableHeader>
		<td> <button id=resetButton onclick="resetGame()"> Reset </button> </td>
		<td> <button id=startButton onclick="startGame()"> Start </button> </td>
		<td> <button id=pauseButton onclick="pauseGame()" disabled> Pause </button> </td>
		<td> <input type=text id=stepCountText width=10px> </td>
		<td> <button id=forwardButton onclick="forwardGame()"> Forward </button> </td>
	</th>
	<tr>
		<td id=errorStatus colSpan=5> </td>
	</tr>
</table>
<table class=tableLayout>
		{{ $board := . }}
		{{ range $row, $cellRow := .Cells }}
		<tr>
			{{ range $col, $dummy := $cellRow }}
			<td id="cell{{$row}}-{{$col}}" {{ if $board.FirstDraw }} onclick="recordActiveCell({{$row}}, {{$col}})" class=newCell {{ else if ($board.IsActive $row $col) }} class=liveCell {{ else }} class=deadCell {{ end }}> {{ if ($board.IsActive $row $col) }} 🌹 {{ end }} </td>
			{{ end }}
		</tr>
		{{ end }}
</table>
`

type uxBoard struct {
	board *Board

	// active cells selected so far. It's only possible to select these cell
	// when the game is not already started.
	activeCells []Cell

	// pauseCh is the channel used to signal the press of pause button
	pauseCh chan struct{}

	// rowXlate is the offset to add to UX coordinates to get the underlying game board's coordinates
	rowXlate int

	// colXlate is the offset to add to UX coordinates to get the underlying game board's coordinates
	colXlate int

	//
	// Following fields are set within drawBoard and used to render the template
	//

	// Cells matrix corresponds to board layout used for ease of rendering the template.
	Cells [][]struct{}

	// FirstDraw is set when board is initialized for the first time
	FirstDraw bool
}

var singleton *uxBoard

func getBoard(new bool) *uxBoard {
	if new {
		singleton = &uxBoard{
			FirstDraw: true,
		}
	}
	return singleton
}

func (b *uxBoard) drawBoard() {
	defer func() {
		b.FirstDraw = false
	}()
	rows := 10
	cols := 10
	rowXlate := 0
	colXlate := 0
	if !b.FirstDraw {
		rows, cols, rowXlate, colXlate = b.computeRowsColsAndXlate(rows, cols)
	}

	b.Cells = make([][]struct{}, rows)
	for i := range b.Cells {
		b.Cells[i] = make([]struct{}, cols)
	}
	b.rowXlate = rowXlate
	b.colXlate = colXlate

	gameSection := js.Global().Get("document").Call("getElementById", "gameSection")
	tmplt, err := template.New("table-template").Parse(boardTmplt)
	if err != nil {
		retError := fmt.Errorf("error parsing template: %w", err)
		gameSection.Set("innerHTML", retError.Error())
		return
	}
	var buf bytes.Buffer
	err = tmplt.Execute(&buf, b)
	if err != nil {
		retError := fmt.Errorf("error executing the template: %w", err)
		gameSection.Set("innerHTML", retError.Error())
		return
	}
	gameSection.Set("innerHTML", buf.String())
}

// computeRowsColsAndXlate computes the number of rows and cols required to plot the board. It also calculates the value to
// convert 0 based row and col values on ux board to corresponding on underlying board values (similar to shifting the origin).
func (b *uxBoard) computeRowsColsAndXlate(minRows, minCols int) (rows, cols, rowXlate, colXlate int) {
	minRow := math.MaxInt64
	maxRow := math.MinInt64
	minCol := math.MaxInt64
	maxCol := math.MinInt64
	for c := range b.board.ActiveCells {
		minRow = int(math.Min(float64(minRow), float64(c.Row)))
		maxRow = int(math.Max(float64(maxRow), float64(c.Row)))
		minCol = int(math.Min(float64(minCol), float64(c.Col)))
		maxCol = int(math.Max(float64(maxCol), float64(c.Col)))
	}
	// count of rows and cols to plot all active cells
	rows = int(math.Max(float64(minRows), float64(maxRow-minRow+1)))
	cols = int(math.Max(float64(minCols), float64(maxCol-minCol+1)))

	minRow -= 2
	minCol -= 2

	// values to convert a point in UX board (0 based) to corresponding values on underlying board
	rowXlate = minRow
	colXlate = minCol

	return
}

const (
	enableResetButton   = 1 << 0
	enableStartButton   = 1 << 1
	enablePauseButton   = 1 << 2
	enableStepCountText = 1 << 3
	enableForwardButton = 1 << 4
)

func (b *uxBoard) setButtonState(stateFlag int) {
	tbl := []struct {
		compName string
		flag     int
	}{
		{"resetButton", enableResetButton},
		{"startButton", enableStartButton},
		{"pauseButton", enablePauseButton},
		{"stepCountText", enableStepCountText},
		{"forwardButton", enableForwardButton},
	}

	for _, e := range tbl {
		disabled := true
		if stateFlag&e.flag > 0 {
			disabled = false
		}
		js.Global().Get("document").Call("getElementById", e.compName).Set("disabled", disabled)
	}
}

func (b *uxBoard) IsActive(row, col int) bool {
	if b.board == nil {
		// when the game board is not yet started the board would be nil
		return false
	}
	cell := Cell{Row: row + b.rowXlate, Col: col + b.colXlate}
	_, ok := b.board.ActiveCells[cell]
	return ok
}

func resetGame(this js.Value, args []js.Value) interface{} {
	b := getBoard(true)
	b.drawBoard()
	return nil
}

func run(stepCount int) {
	b := getBoard(false)
	if b.board == nil {
		b.board = NewBoard(b.activeCells)
	}
	b.pauseCh = make(chan struct{})
	tick := time.NewTicker(time.Second)
loop:
	for {
		select {
		case <-tick.C:
			b.board.Step()
			b.drawBoard()
			b.setButtonState(enablePauseButton)
			if stepCount > 0 {
				stepCount--
				if stepCount == 0 {
					// done with user specified step count
					b.setButtonState(enableResetButton | enableStartButton | enableStepCountText | enableForwardButton)
					if b.pauseCh != nil {
						close(b.pauseCh)
						b.pauseCh = nil
					}
					break loop
				}
			}
		case <-b.pauseCh:
			tick.Stop()
			break loop
		}
	}
}

func startGame(this js.Value, args []js.Value) interface{} {
	go run(-1)
	return nil
}

func forwardGame(this js.Value, args []js.Value) interface{} {
	stepsStr := js.Global().Get("document").Call("getElementById", "stepCountText").Get("value").String()
	steps, err := strconv.Atoi(stepsStr)
	if err != nil {
		js.Global().Get("document").Call("getElementById", "errorStatus").Set("innerHTML", err.Error())
		return nil
	}
	js.Global().Get("document").Call("getElementById", "errorStatus").Set("innerHTML", "")
	go run(steps)
	return nil
}

func pauseGame(this js.Value, args []js.Value) interface{} {
	b := getBoard(false)
	if b.pauseCh != nil {
		close(b.pauseCh)
		b.pauseCh = nil
	}
	b.drawBoard()
	b.setButtonState(enableResetButton | enableStartButton | enableStepCountText | enableForwardButton)
	return nil
}

func recordActiveCell(this js.Value, args []js.Value) interface{} {
	row := args[0].Int()
	col := args[1].Int()
	cell := js.Global().Get("document").Call("getElementById", fmt.Sprintf("cell%d-%d", row, col))
	b := getBoard(false)
	clear := false
	var idx int
	for idx = range b.activeCells {
		if b.activeCells[idx].Row == row && b.activeCells[idx].Col == col {
			clear = true
			break
		}
	}
	if !clear {
		b.activeCells = append(b.activeCells, Cell{Row: row, Col: col})
		cell.Set("className", "liveCell")
	} else {
		b.activeCells = append(b.activeCells[:idx], b.activeCells[idx+1:]...)
		cell.Set("className", "newCell")
	}
	return nil
}

func registerCallbacks() {
	js.Global().Set("resetGame", js.FuncOf(resetGame))
	js.Global().Set("startGame", js.FuncOf(startGame))
	js.Global().Set("forwardGame", js.FuncOf(forwardGame))
	js.Global().Set("pauseGame", js.FuncOf(pauseGame))
	js.Global().Set("recordActiveCell", js.FuncOf(recordActiveCell))
}
