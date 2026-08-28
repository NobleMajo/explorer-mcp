package workspacescan

// IsOutputDir reports whether name is a build/output directory collapsed by default.
func IsOutputDir(name string) bool {
	switch name {
	case "dist", "out", "output":
		return true
	default:
		return false
	}
}

// IsDepsDir reports whether name is a dependency directory collapsed by default.
func IsDepsDir(name string) bool {
	switch name {
	case "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func (o Options) shouldCollapseDir(name string) bool {
	if IsOutputDir(name) && !o.ExpandOutDirs {
		return true
	}
	if IsDepsDir(name) && !o.ExpandDepsDirs {
		return true
	}
	return false
}

func (o Options) truncationSuffix() string {
	if o.TruncateMarker == "" {
		return "/**"
	}
	return o.TruncateMarker
}
