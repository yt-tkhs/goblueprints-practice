package main

import "strings"

const PathSeparator = "/"

type Path struct {
	Path string
	ID string
}

func NewPath(p string) *Path {
	var id string

	p = strings.Trim(p, PathSeparator)      // "/user/1/book/3/" -> "user/1/book/3"
	s := strings.Split(p, PathSeparator)    // "user/1/book/3"   -> [user 1 book 3]

	if len(s) > 1 {
		id = s[len(s) - 1]  // [user 1 book 3] -> 3
		p = strings.Join(s[:len(s) - 1], PathSeparator) // [user 1 book 3] -> "user/1/book"
	}
	return &Path{
		Path:   p,
		ID:     id,
	}
}

func (p *Path) HasID() bool {
	return len(p.ID) > 0
}