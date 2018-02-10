package controllers

import "fmt"

type toFloat float64

func (n toFloat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", n)), nil
}