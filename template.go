// This file is part of Gompet - Copyright 2019-2020 Jari Karjala - www.jpkware.com
// SPDX-License-Identifier: GPLv3-only

package gompet

import (
	"strings"
	"unicode"
)

// VarTemplate holds the pieces of the template in optimal format
// This is not safe for concurrent use!
type VarTemplate struct {
	Pieces  []string        // fixed pieces
	Indices []int           // arg indices for variables between pieces
	Builder strings.Builder // builder for Expand functions
}

// Parse parses the given string to an efficient template
func Parse(str string) *VarTemplate {
	template := VarTemplate{Pieces: make([]string, 0), Indices: make([]int, 0)}
	if str == "" {
		return nil
	}
	split := strings.Split(str, "$")
	var fixedPiece string
	for i, piece := range split {
		if len(piece) > 0 && unicode.IsDigit(rune(piece[0])) {
			template.Pieces = append(template.Pieces, fixedPiece)
			fixedPiece = piece[1:]
			template.Indices = append(template.Indices, int(piece[0])-'0')
		} else {
			if i > 0 {
				fixedPiece += "$" + piece
			} else {
				fixedPiece = piece
			}
		}
	}
	if fixedPiece != "" {
		template.Pieces = append(template.Pieces, fixedPiece)
	}
	return &template
}

// Expand constructs string with variables replaced with given arguments
func (t *VarTemplate) Expand(args []string) string {
	t.Builder.Reset()
	for i, piece := range t.Pieces {
		t.Builder.WriteString(piece)
		if i < len(t.Indices) && t.Indices[i]-1 < len(args) {
			t.Builder.WriteString(args[t.Indices[i]-1])
		}
	}
	return t.Builder.String()
}
