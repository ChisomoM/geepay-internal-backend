package MODULENAME

// public.go exposes the public interfaces that other modules or external code can import and depend on.
// This prevents circular dependencies and clearly defines the module's contract.

// Exported interfaces
var (
	_ Service = (*service)(nil)
	_ Handler = (*handler)(nil)
)
