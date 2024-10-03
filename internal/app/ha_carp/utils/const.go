package CarpUtils

// CARP State Constants
const (
	STAGE_INIT   = "INIT"
	STAGE_BACKUP = "BACKUP"
	STAGE_MASTER = "MASTER"
)

// Log File Constants
const (
	LOG_FOLDER = "/opt/hoster-core/logs"
	LOG_FILE   = LOG_FOLDER + "/ha_carp.log"
)

// Socket File Constants
const SOCKET_FILE = "/var/run/ha_carp.sock"
