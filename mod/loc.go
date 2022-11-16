package mod

import "strings"

type Loc struct {
	URL   string
	proto int
	frag  int
}

func ParseLoc(url string) *Loc {
	l := Loc{URL: url}
	l.init()
	return &l
}

func (l *Loc) init() {
	if l.frag > 0 {
		return
	}
	idx := strings.IndexByte(l.URL, ':')
	if idx > 0 {
		l.proto = idx
	}
	idx = strings.IndexByte(l.URL, '#')
	if idx > 0 {
		l.frag = idx
	} else {
		l.frag = len(l.URL)
	}
}

func (l *Loc) Proto() string {
	if l == nil {
		return ""
	}
	l.init()
	return l.URL[:l.proto]
}
func (l *Loc) Frag() string {
	if l == nil {
		return ""
	}
	l.init()
	if l.frag < len(l.URL) {
		return l.URL[l.frag+1:]
	}
	return ""
}
func (l *Loc) Path() string {
	if l == nil {
		return ""
	}
	l.init()
	p := l.URL
	if l.frag < len(l.URL) {
		p = p[:l.frag]
	}
	if l.proto > 0 {
		p = p[l.proto+1:]
	}
	return p
}

func (l Loc) String() string {
	return l.URL
}
