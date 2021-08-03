package main

import "fmt"

type NaiveColor struct {
	color string
	endl  string
}

func NewColor(color string) *NaiveColor {
	n := new(NaiveColor)
	n.SetColor(color)
	return n
}

func (n *NaiveColor) SetColor(color string) {
	n.color = color
	n.endl = "\033[0m"
}

func (n *NaiveColor) Println(a ...interface{}) {
	fmt.Print(n.color)
	fmt.Print(a...)
	fmt.Print(n.endl)
	fmt.Println()
}

func (n *NaiveColor) Printf(format string, a ...interface{}) {
	fmt.Print(n.color)
	fmt.Printf(format, a...)
	fmt.Print(n.endl)
}
