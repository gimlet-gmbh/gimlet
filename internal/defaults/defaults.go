package defaults

const (
	VERSION           = "00.08.00"
	CODE              = "ctrl"
	CORE_PATH_MAC     = "/usr/local/bin/gmbhCore"
	CORE_PATH_LINUX   = ""
	CORE_PATH_WINDOWS = ""
	CLI_PROMPT        = "[cli] "
	DEFAULT_PROMPT    = "[gmbh] "
)

// For use with UserConfig
const (
	PROJECT_NAME        = "default"
	PROJECT_CONFIG_FILE = "gmbh.yaml"
	DAEMON              = false
	VERBOSE             = true
	DEFAULT_HOST        = "localhost"
	DEFAULT_PORT        = ":59999"
	CONTROL_HOST        = "localhost"
	CONTROL_PORT        = ":59997"
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
	SERVICE_LOG_FILE = ".gmbh"
)

// For use with router
const (
	BASE_ADDRESS = "localhost"
	BASE_PORT    = 49999
)

// For use with process manager
const (
	NUM_RETRIES = 3
)
