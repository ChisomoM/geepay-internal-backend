package auth

// public.go exposes interfaces for this module.
var (
	_ Service = (*service)(nil)
	_ Handler = (*handler)(nil)
)
