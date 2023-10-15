package internal

const (
	NUMBER_OF_NODES                            = 3
	STARTING_PORT                              = 8000
	DELIMETER                                  = ":"
	CLIENT_REQUEST_PREFIX                      = "client_request"
	SERVER_RESPONSE_PREFIX                     = "server_response"
	SERVER_RESPONSE_NON_NUMERIC_REQUEST_NUMBER = "non_numeric_request_number"
	SERVER_RESPONSE_INVALID_REQUEST_NUMER      = "invalid_request_number"
	PREPARE_REQUEST_PREFIX                     = "prepare_request"
	PREPARE_RESPONSE_PREFIX                    = "prepare_response"
	COMMIT_MESSAGE_PREFIX                      = "commit_message"
	INVALID_DATABASE_REQUEST                   = "invalid_database_request"
	VALUE_DOES_NOT_EXIST                       = "value_does_not_exist"
	UPDATE_PERFORMED_SUCCESSFULLY              = "update_performed_successfully"
	MIN_TIMEOUT                                = 10001
	MAX_TIMEOUT                                = 20000
)
