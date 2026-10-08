package formatter

import (
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
)

func (s *formatterState) targetFQN(target ast.TargetRef) string {
	var sb strings.Builder
	for _, pkg := range target.PackagePath {
		sb.WriteString(pkg)
		sb.WriteString(s.packageSeparator)
	}
	sb.WriteString(target.Entity)
	if target.Member != "" {
		sb.WriteString("::")
		sb.WriteString(target.Member)
	}
	return sb.String()
}

func (s *formatterState) formatBlockStatements(stmts []ast.Statement) {
	s.onNewLevel(func() {
		for _, stmt := range stmts {
			formatNode(s, stmt)
		}
	})
}

func (s *formatterState) formatBlockMembers(members []ast.Member) {
	s.onNewLevel(func() {
		for _, member := range members {
			formatNode(s, member)
		}
	})
}
