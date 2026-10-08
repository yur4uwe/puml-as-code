package formatter

import "strings"

func (s *formatterState) getIndent() int {
	return s.indent * s.tabDepth
}

func (s *formatterState) writeIndent() {
	s.buf.WriteString(strings.Repeat(" ", s.indent*s.tabDepth))
}

func (s *formatterState) emitWithIndent(str string) {
	if str != "" && str != "\n" {
		s.writeIndent()
	}
	s.buf.WriteString(str)
}

func (s *formatterState) onNewLevel(f func()) {
	s.tabDepth++
	defer func() { s.tabDepth-- }()
	f()
}
