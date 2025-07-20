package client

import (
	"fmt"
	"strings"
)

type ProgressBar struct {
	total  int64
	length int
	text   string
}

func NewProgressBar(text string, total int64) *ProgressBar {
	return &ProgressBar{text: text, length: 20, total: total}
}

func (p ProgressBar) Set(value int64) {
	value = max(0, min(p.total, value))
	progress := int(float32(value) / float32(p.total) * float32(p.length))
	fmt.Print("\r" + p.text + ": [" +
		strings.Repeat("#", progress) +
		strings.Repeat("-", p.length-progress) +
		"]")
}

func (p ProgressBar) End() {
	fmt.Print(" Done\n")
}
