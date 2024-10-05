package options

type JXLOptions struct {
	debug           bool
	ParseOnly       bool
	RenderVarblocks bool
}

func NewJXLOptions(options *JXLOptions) *JXLOptions {

	opt := &JXLOptions{}
	if options != nil {
		opt.debug = options.debug
		opt.ParseOnly = options.ParseOnly
	}
	return opt
}
