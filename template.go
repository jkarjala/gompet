package gomb

import (
	"strings"
	"unicode"
)

// VarTemplate holds the pieces of the template in optimal format
// This is not safe for concurrent use!
type VarTemplate struct {
	Pieces  []string        // fixed pieces
	Indices []int           // arg indices for variables between pieces
	builder strings.Builder // internal builder
}

// NewVarTemplate parses the given string to an efficient template
func NewVarTemplate(str string) *VarTemplate {
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
	t.builder.Reset()
	for i, piece := range t.Pieces {
		t.builder.WriteString(piece)
		if i < len(t.Indices) && t.Indices[i]-1 < len(args) {
			t.builder.WriteString(args[t.Indices[i]-1])
		}
	}
	return t.builder.String()
}
