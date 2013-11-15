package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

type Column struct {
	Name  string
	Start int
	Width int
}

type Columns []Column

type Row []string

type Rows []Row

type Table struct {
	Columns Columns
	Rows    Rows
	Footer  Row
}

func CreateTable() Table {
	return Table{Columns: make(Columns, 0), Rows: make(Rows, 0)}
}

func (this *Columns) Append(name string, width int) Columns {
	cel := Column{Name: name, Start: 0, Width: width}

	cels := *this
	length := len(cels)

	if length > 0 {
		lastCel := cels[length-1]
		cel.Start = lastCel.Start + lastCel.Width
	}
	return append(cels, cel)
}

func (this *Columns) Println(col int, row int, width int) {
	for _, cel := range *this {
		columnWidth := (cel.Width * width / 100) - 1
		columnSkip := cel.Start * width / 100
		s := cel.Name
		if len(s) > columnWidth {

			for {
				if len(s+"...") <= columnWidth {
					break
				}
				s = s[:len(s)-1]
			}
			s += "..."
		}
		s = "|" + s
		x := col + columnSkip
		for _, r := range s {
			termbox.SetCell(x, row, r, termbox.ColorDefault, termbox.ColorDefault)
			x++
		}
	}
}

func (this *Columns) PrintRowDelimeter(col int, row int, width int) {
	for _, cel := range *this {
		columnWidth := (cel.Width * width / 100)
		columnSkip := cel.Start * width / 100
		for i := col + columnSkip; i <= col+columnSkip+columnWidth; i++ {
			termbox.SetCell(i, row, '-', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (this *Row) Println(col int, row int, width int, columns Columns) {
	for index, value := range *this {
		cel := columns[index]
		columnWidth := (cel.Width * width / 100) - 1
		columnSkip := cel.Start * width / 100
		s := value
		if len(s) > columnWidth {

			for {
				if len(s+"...") <= columnWidth {
					break
				}
				s = s[:len(s)-1]
			}
			s += "..."
		}
		s = "|" + s
		x := col + columnSkip
		for _, r := range s {
			termbox.SetCell(x, row, r, termbox.ColorDefault, termbox.ColorDefault)
			x++
		}
	}
}

func (this *Table) AddColumn(name string, width int) {
	this.Columns = this.Columns.Append(name, width)
}

func (this *Table) AddRow(data ...interface{}) {
	strings := make([]string, len(data))
	for i, d := range data {
		strings[i] = fmt.Sprintf("%v", d)
	}
	this.Rows = append(this.Rows, strings)
}

func (this *Table) SetFooter(data ...interface{}) {
	strings := make([]string, len(data))
	for i, d := range data {
		strings[i] = fmt.Sprintf("%v", d)
	}
	this.Footer = strings
}

func (this *Table) RemoveRows() {
	this.Rows = make(Rows, 0)
}

func (this *Table) Print(col int, row int, width int, height int) {
	currentRow := row
	this.Columns.Println(col, currentRow, width)
	this.Columns.PrintRowDelimeter(col, currentRow+1, width)
	currentRow += 2
	for _, Row := range this.Rows {
		Row.Println(col, currentRow, width, this.Columns)
		currentRow++
		if currentRow+3 >= height {
			break
		}
	}
	this.Columns.PrintRowDelimeter(col, currentRow, width)
	this.Footer.Println(col, currentRow+1, width, this.Columns)
	this.Columns.PrintRowDelimeter(col, currentRow+2, width)
}
