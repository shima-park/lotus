package server

var (
	defaultOptions = Options{
		HTTPAddr: ":8080",
	}
)

type Options struct {
	HTTPAddr     string
	MetadataPath string
	Version      string
	Branch       string
	Commit       string
	Built        string
}

type Option func(*Options)

func HTTPAddr(addr string) Option {
	return func(o *Options) {
		o.HTTPAddr = addr
	}
}

func MetadataPath(path string) Option {
	return func(o *Options) {
		o.MetadataPath = path
	}
}

func Version(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func Branch(branch string) Option {
	return func(o *Options) {
		o.Branch = branch
	}
}

func Commit(commit string) Option {
	return func(o *Options) {
		o.Commit = commit
	}
}

func Built(built string) Option {
	return func(o *Options) {
		o.Built = built
	}
}
