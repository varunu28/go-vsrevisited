package internal

const (
	// cluster configuration
	NUMBER_OF_NODES = 3
	STARTING_PORT   = 8000

	// constants for request
	DELIMETER                                  = ":"
	LOG_DELIMETER                              = "-"
	CLIENT_REQUEST_PREFIX                      = "client_request"
	SERVER_RESPONSE_PREFIX                     = "server_response"
	SERVER_RESPONSE_NON_NUMERIC_REQUEST_NUMBER = "non_numeric_request_number"
	SERVER_RESPONSE_INVALID_REQUEST_NUMER      = "invalid_request_number"
	PREPARE_REQUEST_PREFIX                     = "prepare_request"
	PREPARE_RESPONSE_PREFIX                    = "prepare_response"
	COMMIT_MESSAGE_PREFIX                      = "commit_message"
	CATCHUP_REQUEST_PREFIX                     = "catchup_request"
	CATCHUP_RESPONSE_PREFIX                    = "catchup_response"
	START_VIEW_CHANGE_PREFIX                   = "start_view_change"
	DO_VIEW_CHANGE_PREFIX                      = "do_view_change"
	START_VIEW_PREFIX                          = "start_view"

	// database operation status
	INVALID_DATABASE_REQUEST      = "invalid_database_request"
	VALUE_DOES_NOT_EXIST          = "value_does_not_exist"
	UPDATE_PERFORMED_SUCCESSFULLY = "update_performed_successfully"

	// timeout values associated with server timeout
	MIN_TIMEOUT = 5001
	MAX_TIMEOUT = 20000

	// server states
	NORMAL      = "normal"
	VIEW_CHANGE = "view change"
	RECOVERING  = "recovering"
)
