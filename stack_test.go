package main

import "testing"

func TestStack(t *testing.T) {
	s := stack[token]{}

	tk := s.pop()
	if tk != tokenNoop {
		t.Error("expected empty value from an empty stack")
		return
	}

	tk1 := token{strValue: "1"}
	s.push(tk1)

	if tk := s.pop(); tk != tk1 {
		t.Errorf("expected %+v, but got %+v", tk, tk1)
		return
	}

	tk2 := token{strValue: "2"}
	tk3 := token{strValue: "3"}
	s.push(tk1)
	s.push(tk3)
	s.push(tk2)

	if tk := s.pop(); tk != tk2 {
		t.Errorf("expected %+v, but got %+v", tk, tk2)
		return
	}

	if tk := s.pop(); tk != tk3 {
		t.Errorf("expected %+v, but got %+v", tk, tk3)
		return
	}

	if tk := s.pop(); tk != tk1 {
		t.Errorf("expected %+v, but got %+v", tk, tk1)
		return
	}

	if tk := s.pop(); tk != tokenNoop {
		t.Error("expected empty value from an empty stack")
		return
	}
}
