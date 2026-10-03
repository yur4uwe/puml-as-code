package formatter

type FormatOptions struct {
	IndentSize int
	UseTabs    bool
}

func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		IndentSize: 2,
		UseTabs:    false,
	}
}

// ApplyEditorConfig is a no-op extension hook for future .editorconfig
// file discovery and resolution.
func ApplyEditorConfig(opts *FormatOptions) error {
	// TODO: Implement .editorconfig resolution
	// check the pwd for .editorconfig
	return nil
}
