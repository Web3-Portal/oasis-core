package protocol

import (
	"fmt"
	"reflect"

	beacon "github.com/oasisprotocol/oasis-core/go/beacon/api"
	"github.com/oasisprotocol/oasis-core/go/common"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/errors"
	"github.com/oasisprotocol/oasis-core/go/common/sgx/ias"
	"github.com/oasisprotocol/oasis-core/go/common/version"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	roothash "github.com/oasisprotocol/oasis-core/go/roothash/api"
	"github.com/oasisprotocol/oasis-core/go/roothash/api/block"
	"github.com/oasisprotocol/oasis-core/go/roothash/api/commitment"
	"github.com/oasisprotocol/oasis-core/go/roothash/api/message"
	"github.com/oasisprotocol/oasis-core/go/runtime/transaction"
	storage "github.com/oasisprotocol/oasis-core/go/storage/api"
)

// MethodQueryBatchWeightLimits is the name of the runtime batch weight limits query method.
const MethodQueryBatchWeightLimits = "internal.BatchWeightLimits"

// NOTE: Bump RuntimeProtocol version in go/common/version if you
//       change any of the structures below.

// MessageType is a message type.
type MessageType uint8

// String returns a string representation of a message type.
func (m MessageType) String() string {
	switch m {
	case MessageRequest:
		return "request"
	case MessageResponse:
		return "response"
	default:
		return fmt.Sprintf("[malformed: %d]", m)
	}
}

const (
	// Invalid message (should never be seen on the wire).
	MessageInvalid MessageType = 0

	// Request message.
	MessageRequest MessageType = 1

	// Response message.
	MessageResponse MessageType = 2
)

// Message is a protocol message.
type Message struct {
	ID          uint64      `json:"id"`
	MessageType MessageType `json:"message_type"`
	Body        Body        `json:"body"`
}

// Body is a protocol message body.
type Body struct {
	Empty *Empty `json:",omitempty"`
	Error *Error `json:",omitempty"`

	// Runtime interface.
	RuntimeInfoRequest                    *RuntimeInfoRequest                    `json:",omitempty"`
	RuntimeInfoResponse                   *RuntimeInfoResponse                   `json:",omitempty"`
	RuntimePingRequest                    *Empty                                 `json:",omitempty"`
	RuntimeShutdownRequest                *Empty                                 `json:",omitempty"`
	RuntimeCapabilityTEERakInitRequest    *RuntimeCapabilityTEERakInitRequest    `json:",omitempty"`
	RuntimeCapabilityTEERakInitResponse   *Empty                                 `json:",omitempty"`
	RuntimeCapabilityTEERakReportRequest  *Empty                                 `json:",omitempty"`
	RuntimeCapabilityTEERakReportResponse *RuntimeCapabilityTEERakReportResponse `json:",omitempty"`
	RuntimeCapabilityTEERakAvrRequest     *RuntimeCapabilityTEERakAvrRequest     `json:",omitempty"`
	RuntimeCapabilityTEERakAvrResponse    *Empty                                 `json:",omitempty"`
	RuntimeRPCCallRequest                 *RuntimeRPCCallRequest                 `json:",omitempty"`
	RuntimeRPCCallResponse                *RuntimeRPCCallResponse                `json:",omitempty"`
	RuntimeLocalRPCCallRequest            *RuntimeLocalRPCCallRequest            `json:",omitempty"`
	RuntimeLocalRPCCallResponse           *RuntimeLocalRPCCallResponse           `json:",omitempty"`
	RuntimeCheckTxBatchRequest            *RuntimeCheckTxBatchRequest            `json:",omitempty"`
	RuntimeCheckTxBatchResponse           *RuntimeCheckTxBatchResponse           `json:",omitempty"`
	RuntimeExecuteTxBatchRequest          *RuntimeExecuteTxBatchRequest          `json:",omitempty"`
	RuntimeExecuteTxBatchResponse         *RuntimeExecuteTxBatchResponse         `json:",omitempty"`
	RuntimeAbortRequest                   *Empty                                 `json:",omitempty"`
	RuntimeAbortResponse                  *Empty                                 `json:",omitempty"`
	RuntimeKeyManagerPolicyUpdateRequest  *RuntimeKeyManagerPolicyUpdateRequest  `json:",omitempty"`
	RuntimeKeyManagerPolicyUpdateResponse *Empty                                 `json:",omitempty"`
	RuntimeQueryRequest                   *RuntimeQueryRequest                   `json:",omitempty"`
	RuntimeQueryResponse                  *RuntimeQueryResponse                  `json:",omitempty"`
	RuntimeConsensusSyncRequest           *RuntimeConsensusSyncRequest           `json:",omitempty"`
	RuntimeConsensusSyncResponse          *Empty                                 `json:",omitempty"`

	// Host interface.
	HostRPCCallRequest              *HostRPCCallRequest              `json:",omitempty"`
	HostRPCCallResponse             *HostRPCCallResponse             `json:",omitempty"`
	HostStorageSyncRequest          *HostStorageSyncRequest          `json:",omitempty"`
	HostStorageSyncResponse         *HostStorageSyncResponse         `json:",omitempty"`
	HostLocalStorageGetRequest      *HostLocalStorageGetRequest      `json:",omitempty"`
	HostLocalStorageGetResponse     *HostLocalStorageGetResponse     `json:",omitempty"`
	HostLocalStorageSetRequest      *HostLocalStorageSetRequest      `json:",omitempty"`
	HostLocalStorageSetResponse     *Empty                           `json:",omitempty"`
	HostFetchConsensusBlockRequest  *HostFetchConsensusBlockRequest  `json:",omitempty"`
	HostFetchConsensusBlockResponse *HostFetchConsensusBlockResponse `json:",omitempty"`
	HostFetchTxBatchRequest         *HostFetchTxBatchRequest         `json:",omitempty"`
	HostFetchTxBatchResponse        *HostFetchTxBatchResponse        `json:",omitempty"`
}

// Type returns the message type by determining the name of the first non-nil member.
func (body Body) Type() string {
	b := reflect.ValueOf(body)
	for i := 0; i < b.NumField(); i++ {
		if !b.Field(i).IsNil() {
			return reflect.TypeOf(body).Field(i).Name
		}
	}
	return ""
}

// Empty is an empty message body.
type Empty struct{}

// Error is a message body representing an error.
type Error struct {
	Module  string `json:"module,omitempty"`
	Code    uint32 `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// String returns a string representation of this runtime error.
func (e Error) String() string {
	return fmt.Sprintf("runtime error: module: %s code: %d message: %s", e.Module, e.Code, e.Message)
}

// RuntimeInfoRequest is a worker info request message body.
type RuntimeInfoRequest struct {
	// RuntimeID is the assigned runtime ID of the loaded runtime.
	RuntimeID common.Namespace `json:"runtime_id"`

	// ConsensusBackend is the name of the consensus backend that is in use for the consensus layer.
	ConsensusBackend string `json:"consensus_backend"`
	// ConsensusProtocolVersion is the consensus protocol version that is in use for the consensus
	// layer.
	ConsensusProtocolVersion version.Version `json:"consensus_protocol_version"`
	// ConsensusChainContext is the consensus layer chain domain separation context.
	ConsensusChainContext string `json:"consensus_chain_context"`

	// LocalConfig is the node-local runtime configuration.
	//
	// This configuration must not be used in any context which requires determinism across
	// replicated runtime instances.
	LocalConfig map[string]interface{} `json:"local_config,omitempty"`
}

// Feature is a set of supported runtime features.
type Features struct {
	// ScheduleControl is the schedule control feature.
	ScheduleControl *FeatureScheduleControl `json:"schedule_control,omitempty"`
}

// HasScheduleControl returns true when the runtime supports the schedule control feature.
func (f *Features) HasScheduleControl() bool {
	return f != nil && f.ScheduleControl != nil
}

// FeatureScheduleControl is a feature specifying that the runtime supports controlling the
// scheduling of batches. This means that the scheduler should only take priority into account and
// ignore weights, leaving it up to the runtime to decide which transactions to include.
type FeatureScheduleControl struct {
	// InitialBatchSize is the size of the initial batch of transactions.
	InitialBatchSize uint32 `json:"initial_batch_size"`
}

// RuntimeInfoResponse is a runtime info response message body.
type RuntimeInfoResponse struct {
	// ProtocolVersion is the runtime protocol version supported by the runtime.
	ProtocolVersion version.Version `json:"protocol_version"`

	// RuntimeVersion is the version of the runtime.
	RuntimeVersion version.Version `json:"runtime_version"`

	// Features describe the features supported by the runtime.
	Features *Features `json:"features,omitempty"`
}

// RuntimeCapabilityTEERakInitRequest is a worker RFC 0009 CapabilityTEE
// initialization request message body.
type RuntimeCapabilityTEERakInitRequest struct {
	TargetInfo []byte `json:"target_info"`
}

// RuntimeCapabilityTEERakReportResponse is a worker RFC 0009 CapabilityTEE RAK response message body.
type RuntimeCapabilityTEERakReportResponse struct {
	RakPub signature.PublicKey `json:"rak_pub"`
	Report []byte              `json:"report"`
	Nonce  string              `json:"nonce"`
}

// RuntimeCapabilityTEERakAvrRequest is a worker RFC 0009 CapabilityTEE RAK AVR setup request message body.
type RuntimeCapabilityTEERakAvrRequest struct {
	AVR ias.AVRBundle `json:"avr"`
}

// RuntimeRPCCallRequest is a worker RPC call request message body.
type RuntimeRPCCallRequest struct {
	// Request.
	Request []byte `json:"request"`
}

// RuntimeRPCCallResponse is a worker RPC call response message body.
type RuntimeRPCCallResponse struct {
	// Response.
	Response []byte `json:"response"`
}

// RuntimeLocalRPCCallRequest is a worker local RPC call request message body.
type RuntimeLocalRPCCallRequest struct {
	// Request.
	Request []byte `json:"request"`
}

// RuntimeLocalRPCCallResponse is a worker local RPC call response message body.
type RuntimeLocalRPCCallResponse struct {
	// Response.
	Response []byte `json:"response"`
}

// RuntimeCheckTxBatchRequest is a worker check tx batch request message body.
type RuntimeCheckTxBatchRequest struct {
	// ConsensusBlock is the consensus light block at the last finalized round
	// height (e.g., corresponding to .Block.Header.Round).
	ConsensusBlock consensus.LightBlock `json:"consensus_block"`

	// Batch of runtime inputs to check.
	Inputs transaction.RawBatch `json:"inputs"`
	// Block on which the batch check should be based.
	Block block.Block `json:"block"`
	// Epoch is the current epoch number.
	Epoch beacon.EpochTime `json:"epoch"`

	// MaxMessages is the maximum number of messages that can be emitted in this
	// round. Any more messages will be rejected by the consensus layer.
	MaxMessages uint32 `json:"max_messages"`
}

// CheckTxResult contains the result of a CheckTx operation.
type CheckTxResult struct {
	// Error is the error (if any) that resulted from the operation.
	Error Error `json:"error"`

	// Meta contains metadata that can be used for scheduling transactions by the scheduler.
	Meta *CheckTxMetadata `json:"meta,omitempty"`
}

// CheckTxMetadata is the transaction check-tx metadata.
type CheckTxMetadata struct {
	// Priority is the transaction's priority.
	Priority uint64 `json:"priority,omitempty"`

	// Weight are runtime specific transaction weights.
	Weights map[transaction.Weight]uint64 `json:"weights,omitempty"`
}

// IsSuccess returns true if transaction execution was successful.
func (r *CheckTxResult) IsSuccess() bool {
	return r.Error.Code == errors.CodeNoError
}

// ToCheckedTransaction creates CheckedTransaction from CheckTx result.
//
// Assumes a successful result.
func (r *CheckTxResult) ToCheckedTransaction(rawTx []byte) *transaction.CheckedTransaction {
	switch r.Meta {
	case nil:
		return transaction.NewCheckedTransaction(rawTx, 0, nil)
	default:
		return transaction.NewCheckedTransaction(rawTx, r.Meta.Priority, r.Meta.Weights)
	}
}

// RuntimeCheckTxBatchResponse is a worker check tx batch response message body.
type RuntimeCheckTxBatchResponse struct {
	// Batch of CheckTx results corresponding to transactions passed on input.
	Results []CheckTxResult `json:"results"`
}

// ComputedBatch is a computed batch.
type ComputedBatch struct {
	// Header is the compute results header.
	Header commitment.ComputeResultsHeader `json:"header"`
	// Log that generates the I/O tree.
	IOWriteLog storage.WriteLog `json:"io_write_log"`
	// Batch of storage write operations.
	StateWriteLog storage.WriteLog `json:"state_write_log"`
	// If this runtime uses a TEE, then this is the signature of Header with
	// node's RAK for this runtime.
	RakSig signature.RawSignature `json:"rak_sig"`
	// Messages are the emitted runtime messages.
	Messages []message.Message `json:"messages"`
}

// String returns a string representation of a computed batch.
func (b *ComputedBatch) String() string {
	return "<ComputedBatch>"
}

// ExecutionMode is the batch execution mode.
type ExecutionMode uint8

const (
	// ExecutionModeExecute is the execution mode where the batch of transactions is exectued as-is
	// without the ability to perform any modifications to the batch.
	ExecutionModeExecute = 0
	// ExecutionModeSchedule is the execution mode where the runtime is in control of scheduling and
	// may arbitrarily modify the batch during execution.
	//
	// This execution mode will only be used in case the runtime advertises to support the schedule
	// control feature. In this case the call will only contain up to InitialBatchSize transactions
	// and the runtime will need to request more if it needs more.
	ExecutionModeSchedule = 1
)

// RuntimeExecuteTxBatchRequest is a worker execute tx batch request message body.
type RuntimeExecuteTxBatchRequest struct {
	// Mode is the execution mode.
	Mode ExecutionMode `json:"mode,omitempty"`

	// ConsensusBlock is the consensus light block at the last finalized round
	// height (e.g., corresponding to .Block.Header.Round).
	ConsensusBlock consensus.LightBlock `json:"consensus_block"`

	// RoundResults are the results of executing the previous successful round.
	RoundResults *roothash.RoundResults `json:"round_results"`

	// IORoot is the I/O root containing the inputs (transactions) that
	// the compute node should use. It must match what is passed in "inputs".
	IORoot hash.Hash `json:"io_root"`
	// Batch of inputs (transactions).
	Inputs transaction.RawBatch `json:"inputs"`
	// InMessages are the incoming messages emitted by the consensus layer.
	InMessages []*message.IncomingMessage `json:"in_msgs,omitempty"`
	// Block on which the batch computation should be based.
	Block block.Block `json:"block"`
	// Epoch is the current epoch number.
	Epoch beacon.EpochTime `json:"epoch"`

	// MaxMessages is the maximum number of messages that can be emitted in this
	// round. Any more messages will be rejected by the consensus layer.
	MaxMessages uint32 `json:"max_messages"`
}

// RuntimeExecuteTxBatchResponse is a worker execute tx batch response message body.
type RuntimeExecuteTxBatchResponse struct {
	Batch             ComputedBatch                 `json:"batch"`
	BatchWeightLimits map[transaction.Weight]uint64 `json:"batch_weight_limits"`

	// TxHashes are the transaction hashes of the included batch.
	TxHashes []hash.Hash `json:"tx_hashes,omitempty"`
	// TxRejectHashes are the transaction hashes of transactions that should be immediately removed
	// from the scheduling queue as they are invalid.
	TxRejectHashes []hash.Hash `json:"tx_reject_hashes,omitempty"`
	// TxInputRoot is the root hash of all transaction inputs.
	TxInputRoot hash.Hash `json:"tx_input_root,omitempty"`
	// TxInputWriteLog is the write log for generating transaction inputs.
	TxInputWriteLog storage.WriteLog `json:"tx_input_write_log,omitempty"`
}

// RuntimeKeyManagerPolicyUpdateRequest is a runtime key manager policy request
// message body.
type RuntimeKeyManagerPolicyUpdateRequest struct {
	SignedPolicyRaw []byte `json:"signed_policy_raw"`
}

// RuntimeQueryRequest is a runtime query request message body.
type RuntimeQueryRequest struct {
	// ConsensusBlock is the consensus light block at the last finalized round
	// height (e.g., corresponding to .Header.Round).
	ConsensusBlock consensus.LightBlock `json:"consensus_block"`

	// Header is the current block header.
	Header block.Header `json:"header"`
	// Epoch is the current epoch number.
	Epoch beacon.EpochTime `json:"epoch"`

	// MaxMessages is the maximum number of messages that can be emitted in this
	// round. Any more messages will be rejected by the consensus layer.
	MaxMessages uint32 `json:"max_messages"`

	Method string `json:"method"`
	Args   []byte `json:"args,omitempty"`
}

// RuntimeQueryRequest is a runtime query response message body.
type RuntimeQueryResponse struct {
	Data []byte `json:"data,omitempty"`
}

// RuntimeConsensusSyncRequest is a runtime consensus block synchronization request message body.
type RuntimeConsensusSyncRequest struct {
	Height uint64 `json:"height"`
}

// HostRPCCallRequest is a host RPC call request message body.
type HostRPCCallRequest struct {
	Endpoint string `json:"endpoint"`
	Request  []byte `json:"request"`
}

// HostRPCCallResponse is a host RPC call response message body.
type HostRPCCallResponse struct {
	Response []byte `json:"response"`
}

// HostStorageEndpoint is the host storage endpoint.
type HostStorageEndpoint uint8

const (
	// HostStorageEndpointRuntime is the runtime state storage endpoint.
	HostStorageEndpointRuntime HostStorageEndpoint = 0
	// HostStorageEndpointConsensus is the consensus layer state storage endpoint.
	HostStorageEndpointConsensus HostStorageEndpoint = 1
)

// HostStorageSyncRequest is a host storage read syncer request message body.
type HostStorageSyncRequest struct {
	// Endpoint is the storage endpoint to which this request should be routed.
	Endpoint HostStorageEndpoint `json:"endpoint,omitempty"`

	SyncGet         *storage.GetRequest         `json:",omitempty"`
	SyncGetPrefixes *storage.GetPrefixesRequest `json:",omitempty"`
	SyncIterate     *storage.IterateRequest     `json:",omitempty"`
}

// HostStorageSyncResponse is a host storage read syncer response body.
type HostStorageSyncResponse struct {
	ProofResponse *storage.ProofResponse `json:",omitempty"`
}

// HostLocalStorageGetRequest is a host local storage get request message body.
type HostLocalStorageGetRequest struct {
	Key []byte `json:"key"`
}

// HostLocalStorageGetResponse is a host local storage get response message body.
type HostLocalStorageGetResponse struct {
	Value []byte `json:"value"`
}

// HostLocalStorageSetRequest is a host local storage set request message body.
type HostLocalStorageSetRequest struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

// HostFetchConsensusBlockRequest is a request to host to fetch the given consensus light block.
type HostFetchConsensusBlockRequest struct {
	Height uint64 `json:"height"`
}

// HostFetchConsensusBlockResponse is a response from host fetching the given consensus light block.
type HostFetchConsensusBlockResponse struct {
	Block consensus.LightBlock `json:"block"`
}

// HostFetchTxBatchRequest is a request to host to fetch a further batch of transactions. The host
// will reply with the transactions that it hasn't sent yet.
type HostFetchTxBatchRequest struct {
	// Offset specifies the offset of the batch within the transaction pool.
	//
	// Note that this may miss transactions that have arrived after the last batch was fetched.
	Offset uint32 `json:"offset"`
	// Limit specifies the maximum size of the batch to request.
	Limit uint32 `json:"limit"`
}

// HostFetchTxBatchResponse is a response from host fetching the given transaction batch.
type HostFetchTxBatchResponse struct {
	// Batch is a batch of transactions.
	Batch [][]byte `json:"batch,omitempty"`
}
