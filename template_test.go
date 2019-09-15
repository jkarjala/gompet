package gompet

import (
	"reflect"
	"testing"
)

func TestNewTemplate1(t *testing.T) {
	got := Parse("ab$1cd")
	if !reflect.DeepEqual(got.Pieces, []string{"ab", "cd"}) {
		t.Errorf("Pieces = %v", got.Pieces)
	}
	if !reflect.DeepEqual(got.Indices, []int{1}) {
		t.Errorf("Indices = %v", got.Pieces)
	}
}

func TestNewTemplate2(t *testing.T) {
	got := Parse("$1cd")
	if !reflect.DeepEqual(got.Pieces, []string{"", "cd"}) {
		t.Errorf("Pieces = %v", got.Pieces)
	}
	if !reflect.DeepEqual(got.Indices, []int{1}) {
		t.Errorf("Indices = %v", got.Pieces)
	}
}

func TestNewTemplate3(t *testing.T) {
	got := Parse("ab$1")
	if !reflect.DeepEqual(got.Pieces, []string{"ab"}) {
		t.Errorf("Pieces = %v", got.Pieces)
	}
	if !reflect.DeepEqual(got.Indices, []int{1}) {
		t.Errorf("Indices = %v", got.Pieces)
	}
}

func TestNewTemplate4(t *testing.T) {
	got := Parse("$ab$1cd$")
	if !reflect.DeepEqual(got.Pieces, []string{"$ab", "cd$"}) {
		t.Errorf("Pieces = %v", got.Pieces)
	}
	if !reflect.DeepEqual(got.Indices, []int{1}) {
		t.Errorf("Indices = %v", got.Pieces)
	}
}

func TestNewTemplate5(t *testing.T) {
	got := Parse("$ab$1cd$3ef$2")
	if !reflect.DeepEqual(got.Pieces, []string{"$ab", "cd", "ef"}) {
		t.Errorf("Pieces = %v", got.Pieces)
	}
	if !reflect.DeepEqual(got.Indices, []int{1, 3, 2}) {
		t.Errorf("Indices = %v", got.Pieces)
	}
}

func TestExpand1(t *testing.T) {
	temp := Parse("$ab$1cd$")
	got := temp.Expand([]string{"A"})
	if got != "$abAcd$" {
		t.Errorf("got = '%v'", got)
	}
}

func TestExpand2(t *testing.T) {
	temp := Parse("$3ab$1cd$2")
	got := temp.Expand([]string{"A", "B", "C"})
	if got != "CabAcdB" {
		t.Errorf("got = '%v'", got)
	}
}

func TestExpand3(t *testing.T) {
	temp := Parse("$1ab$1cd$2")
	got := temp.Expand([]string{"A"})
	if got != "AabAcd" {
		t.Errorf("got = '%v'", got)
	}
}

func BenchmarkExpand(b *testing.B) {
	temp := Parse("ab $1 cd $2 ef $3")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = temp.Expand([]string{"A", "B", "C"})
	}
}
