package ErrorMappings

// Status represents the enumeration of possible status values.
type Status int

const (
	Unknown Status = iota
	VmDoesntExist
	HostNotFound
	HostIsDisabled
	SnapTypeDoesntExist
	JailDoesntExist
	ResourceDoesntExist
	SnapshotDoesntExist
	CouldNotParseYourInput
)

var statusStrings = map[Status]string{
	VmDoesntExist:          "vm doesn't exist",
	HostNotFound:           "host was not found in our database",
	HostIsDisabled:         "host is disabled",
	SnapTypeDoesntExist:    "snapshot type doesn't exist",
	JailDoesntExist:        "jail doesn't exist",
	ResourceDoesntExist:    "resource doesn't exist",
	SnapshotDoesntExist:    "snapshot doesn't exist",
	CouldNotParseYourInput: "could not parse your input",
}

// String returns the string representation of the Status value.
func (s Status) String() string {
	return statusStrings[s]
}

// Look up a string value and return a corresponding INT value
func ValueLookup(value string) Status {
	for k, v := range statusStrings {
		if v == value {
			return k
		}
	}
	return 0
}

// Example usage
// status := StatusInProgress
// fmt.Printf("Status: %s\n", status.String())

// Lookup string value based on integer value
// lookupStatus := StatusCompleted
// fmt.Printf("Lookup Status: %s\n", statusStrings[lookupStatus])
