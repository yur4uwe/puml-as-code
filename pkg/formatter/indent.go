package formatter

import "strings"

func (s *formatterState) getIndent() int {
	return s.indent * s.tabDepth
}

func (s *formatterState) writeIndent() {
	if s.useTabs {
		s.buf.WriteString(strings.Repeat("\t", s.tabDepth))
		return
	}
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
	savedLast := s.lastNode
	s.lastNode = nil
	defer func() {
		s.tabDepth--
		s.lastNode = savedLast
	}()
	f()
}
