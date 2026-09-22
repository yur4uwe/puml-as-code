package dialect

func Factory(name string) Dialect {
	switch name {
	case "go":
		return NewGoDialect()
	case "lax":
		return LaxDialect{}
	default:
		return nil
	}
}
