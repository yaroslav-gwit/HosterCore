package handlers

// Standard error response. Used in Swagger docs and inside the main error handler.
type SwaggerError struct {
	ErrorID    int    `json:"id"`
	ErrorValue string `json:"message"`
}

// Purely Swagger related object, not used anywhere else in the codebase
type SwaggerSuccess struct {
	Message string `json:"message"` // success
}

// Purely Swagger related object, not used anywhere else in the codebase
type SwaggerStringList struct {
	Message []string `json:"message"` // success
}

type JailCloneInput struct {
	JailName     string `json:"jail_name"`
	NewJailName  string `json:"new_jail_name"`
	SnapshotName string `json:"snapshot_name"`
}

type VmCloneInput struct {
	VmName       string `json:"vm_name"`
	NewVmName    string `json:"new_vm_name"`
	SnapshotName string `json:"snapshot_name"`
}

type VmCpuInput struct {
	CpuSockets int `json:"cpu_sockets"`
	CpuCores   int `json:"cpu_cores"`
	CpuThreads int `json:"cpu_threads"`
}

type VmRamInput struct {
	RamAmount  int    `json:"ram_amount"`
	BytesValue string `json:"bytes_value"`
}

type VmOsSettings struct {
	OsType    string `json:"os_type"`
	OsComment string `json:"os_comment"`
}
