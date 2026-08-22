package dialect

func Factory(name string) Dialect {
	switch name {
	case "go":
		return NewGoDialect()
	default:
		return nil
	}
}

