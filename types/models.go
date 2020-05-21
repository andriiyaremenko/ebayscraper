package types

import (
	_ "encoding/json"
)

type Product struct {
	Title      string            `json:"title"`
	ImageLink  string            `json:"imageLink"`
	Attributes map[string]string `json:"attributes"`
}
