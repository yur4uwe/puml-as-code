package gogenerator

import (
	"fmt"
	"log"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/resolver"
)

func isInterface(ent *ast.Entity) bool {
	if ent == nil {
		return false
	}
	return ent.Kind == ast.EntityInterface ||
		ent.Kind == ast.EntityAbstractClass ||
		ent.Kind == ast.EntityProtocol
}

func isStruct(ent *ast.Entity) bool {
	if ent == nil {
		return true // default case
	}
	return ent.Kind == ast.EntityClass ||
		ent.Kind == ast.EntityStruct ||
		ent.Kind == ast.EntityRecord ||
		ent.Kind == ast.EntityDataclass ||
		ent.Kind == ast.EntityException
}

func isEnum(ent *ast.Entity) bool {
	if ent == nil {
		return false
	}
	return ent.Kind == ast.EntityEnum
}

func tryInterfacePromotion(symb *resolver.EntitySymbol) error {
	if symb.AST == nil {
		symb.AST = &ast.Entity{
			Identifier: simpleName(symb.FQN),
			Kind:       ast.EntityInterface,
		}
		return nil
	}
	for _, member := range symb.AST.Members {
		switch member.(type) {
		case *dialect.GoField:
			return fmt.Errorf("cannot realize class %s as an interface in Go: it contains fields", symb.FQN)
		}
	}
	symb.AST.Kind = ast.EntityInterface
	return nil
}

func simpleName(fqn string) string {
	if idx := strings.LastIndex(fqn, "."); idx != -1 {
		return fqn[idx+1:]
	}
	return fqn
}

func (GoCodeGenerator) SemanticPass(tbl *resolver.SymbolTable) error {
	for _, rel := range tbl.Relationships {
		switch rel.Type {
		case ast.RelationInheritance:
			// Promotion check
			if isInterface(rel.Source.AST) && isStruct(rel.Target.AST) {
				if err := tryInterfacePromotion(rel.Target); err != nil {
					return err
				}
				log.Printf("warning: conditions met for interface promotion, %s will become an interface", rel.Target.FQN)
			} else if isStruct(rel.Source.AST) && isInterface(rel.Target.AST) {
				rel.Type = ast.RelationRealization
			}
		case ast.RelationRealization:
			if isStruct(rel.Target.AST) {
				// Promotion check
				if err := tryInterfacePromotion(rel.Target); err != nil {
					return err
				}
				log.Printf("warning: conditions met for interface promotion, %s will become an interface", rel.Target.FQN)
			} else if isInterface(rel.Source.AST) {
				rel.Type = ast.RelationInheritance
			}
		}
	}

	for _, ent := range tbl.Entities {
		switch {
		case isInterface(ent.AST):
			for _, member := range ent.AST.Members {
				switch member.(type) {
				case *dialect.GoField:
					return fmt.Errorf("interface or abstract class %s cannot declare fields", ent.FQN)
				}
			}
		case isEnum(ent.AST):
			if len(ent.AST.Members) == 0 {
				return fmt.Errorf("enum %s must declare fields", ent.FQN)
			}
		}
	}
	return nil
}
