package defaults

const (
	VERSION           = "0.9.1"
	CODE              = "divorce"
	CORE_PATH_MAC     = "/usr/local/bin/gmbhCore"
	LOG_PATH          = "/gmbh/"
	CORE_LOG_NAME     = "core.log"
	ACCESS_LOG_NAME   = "access.log"
	ERROR_LOG_NAME    = "error.log"
	CORE_PATH_LINUX   = ""
	CORE_PATH_WINDOWS = ""
	CLI_PROMPT        = "[cli] "
	CTRL_PROMPT       = "[gmbh] "
	DEFAULT_PROMPT    = "[gmbh] "
)

// For use with UserConfig
const (
	PROJECT_NAME        = "default"
	PROJECT_CONFIG_FILE = "gmbh.yaml"
	DAEMON              = false
	VERBOSE             = true
	DEFAULT_HOST        = "localhost"
	DEFAULT_PORT        = ":49500"
	CONTROL_HOST        = "localhost"
	CONTROL_PORT        = ":59500"
)

// For use with ServiceConfig
const (
	SERVICE_NAME = "default"
	IS_CLIENT    = true
	IS_SERVER    = true
)

// For use with services
const (
	CONFIG_FILE      = "/gmbh.yaml"
	CONFIG_FILE_EXT  = ".yaml"
	SERVICE_LOG_PATH = "/gmbh/"
	SERVICE_LOG_FILE = "core.log"
)

// For use with router
const (
	BASE_ADDRESS = "localhost"
	BASE_PORT    = 49999
)

// For use with process manager
const (
	STARTING_ID = 100
	NUM_RETRIES = 3
	TIMEOUT     = 30
)

// Addresses
const (
	CORE_ADDRESS = "localhost:49500"
	CORE_START   = 49502

	PM_ADDRESS = "localhost:59500"
	PM_START   = 59502

	LOCALHOST = "localhost"
)
