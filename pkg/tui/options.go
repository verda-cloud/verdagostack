package tui

// ResolveConfirmConfig applies options to a default ConfirmConfig.
func ResolveConfirmConfig(opts []ConfirmOption) ConfirmConfig {
	cfg := ConfirmConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// ResolveTextInputConfig applies options to a default TextInputConfig.
func ResolveTextInputConfig(opts []TextInputOption) TextInputConfig {
	cfg := TextInputConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// ResolveSelectConfig applies options to a default SelectConfig.
func ResolveSelectConfig(opts []SelectOption) SelectConfig {
	cfg := SelectConfig{PageSize: 10, Loop: true}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// ResolveMultiSelectConfig applies options to a default MultiSelectConfig.
func ResolveMultiSelectConfig(opts []MultiSelectOption) MultiSelectConfig {
	cfg := MultiSelectConfig{PageSize: 10, Loop: true}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// ResolveEditorConfig applies options to a default EditorConfig.
func ResolveEditorConfig(opts []EditorOption) EditorConfig {
	cfg := EditorConfig{FileExt: ".txt", ShowHelp: true}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
