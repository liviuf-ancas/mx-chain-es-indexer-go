package indexer

import "errors"

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilAccountsDB signals that a nil accounts database has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrNegativeDenominationValue signals that a negative denomination value has been provided
var ErrNegativeDenominationValue = errors.New("negative denomination value")

// ErrNilDataDispatcher signals that an operation has been attempted to or with a nil data dispatcher implementation
var ErrNilDataDispatcher = errors.New("nil data dispatcher")

// ErrNilElasticProcessor signals that an operation has been attempted to or with a nil elastic processor implementation
var ErrNilElasticProcessor = errors.New("nil elastic processor")

// ErrNegativeCacheSize signals that an invalid cache size has been provided
var ErrNegativeCacheSize = errors.New("negative cache size")

// ErrEmptyEnabledIndexes signals that an empty slice of enables indexes has been provided
var ErrEmptyEnabledIndexes = errors.New("empty enabled indexes slice")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrCouldNotCreatePolicy signals that the index policy hasn't been created
var ErrCouldNotCreatePolicy = errors.New("could not create policy")

// ErrNoElasticUrlProvided signals that the url to the elasticsearch database hasn't been provided
var ErrNoElasticUrlProvided = errors.New("no elastic url provided")

// ErrBackOff signals that an error was received from the server
var ErrBackOff = errors.New("back off something is not working well")

// ErrNilTransactionFeeCalculator signals that a nil transaction fee calculator has been provided
var ErrNilTransactionFeeCalculator = errors.New("nil transaction fee calculator")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilUrl signals that the provided url is empty
var ErrNilUrl = errors.New("url is empty")

// ErrCannotCastAccountHandlerToUserAccount signals that cast from AccountHandler to UserAccountHandler was not OK
var ErrCannotCastAccountHandlerToUserAccount = errors.New("cannot cast AccountHandler to UserAccountHandler")

// ErrNilHeaderHandler signals that a nil header handler has been provided
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNilBlockBody signals that a nil block body has been provided
var ErrNilBlockBody = errors.New("nil block body")

// ErrHeaderTypeAssertion signals that body type assertion failed
var ErrHeaderTypeAssertion = errors.New("header type assertion failed")

// ErrNilElasticBlock signals that a nil elastic block has been provided
var ErrNilElasticBlock = errors.New("nil elastic block")

// ErrNilElasticProcessorArguments signals that a nil arguments for elastic processor has been provided
var ErrNilElasticProcessorArguments = errors.New("nil elastic processor arguments")

// ErrNilEnabledIndexesMap signals that a nil enabled indexes map has been provided
var ErrNilEnabledIndexesMap = errors.New("nil enabled indexes map")

// ErrNilDatabaseClient signals that a nil database client has been provided
var ErrNilDatabaseClient = errors.New("nil database client")

// ErrNilStatisticHandler signals that a nil statistics handler has been provided
var ErrNilStatisticHandler = errors.New("nil statistics handler")

// ErrNilBlockHandler signals that a nil block handler has been provided
var ErrNilBlockHandler = errors.New("nil block handler")

// ErrNilAccountsHandler signals that a nil accounts handler has been provided
var ErrNilAccountsHandler = errors.New("nil accounts handler")

// ErrNilMiniblocksHandler signals that a nil miniblocks handler has been provided
var ErrNilMiniblocksHandler = errors.New("nil miniblocks handler")

// ErrNilValidatorsHandler signals that a nil validators handler has been provided
var ErrNilValidatorsHandler = errors.New("nil validators handler")

// ErrNilTransactionsHandler signals that a nil transactions handler has been provided
var ErrNilTransactionsHandler = errors.New("nil transactions handler")

// ErrNilTransactionsProcessorArguments signals that a nil arguments structure for transactions processor has been provided
var ErrNilTransactionsProcessorArguments = errors.New("nil transactions processor args")

// ErrNilPool signals that a nil transaction pool has been provided
var ErrNilPool = errors.New("nil transaction pool")

// ErrNilLogsAndEventsHandler signals that a nil logs and events handler has been provided
var ErrNilLogsAndEventsHandler = errors.New("nil logs and events handler")

// ErrNilBalanceConverter signals that a nil balance converter has been provided
var ErrNilBalanceConverter = errors.New("nil balance converter")

// ErrNilOperationsHandler signals that a nil operations handler has been provided
var ErrNilOperationsHandler = errors.New("nil operations handler")
