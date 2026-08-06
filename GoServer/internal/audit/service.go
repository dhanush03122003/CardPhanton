package audit

// Service is reserved for audit domain use cases that sit above the repository layer.
// The current codebase only requires repository-level persistence, so this stays as a thin seam.
type Service interface{}
