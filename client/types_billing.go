package client

// This file defines the Go structs corresponding to the JSON request and response
// schemas for the Billing and Credits related APIs (/v1/credits, /v1/credits/coinbase).

// --- Structs for GET /credits ---

// CreditsData represents the core credit balance information for an account.
type CreditsData struct {
	TotalCredits float64 `json:"total_credits"` // Total credits purchased in USD.
	TotalUsage   float64 `json:"total_usage"`   // Total credits used in USD.
}

// Note: The response structure for GET /credits is defined in billing.go
// as:
// type CreditsGetResponse struct {
//	 Data CreditsData `json:"data"`
// }

// --- Structs for POST /credits/coinbase ---

// CoinbaseChargeRequest represents the request body for creating a Coinbase Commerce charge.
type CoinbaseChargeRequest struct {
	Amount  float64 `json:"amount"`   // Required: USD amount to charge (must be between min and max purchase limits).
	Sender  string  `json:"sender"`   // Required: Ethereum address (hex format) of the sender.
	ChainID int     `json:"chain_id"` // Required: EVM Chain ID for the transaction (e.g., 1, 137, 8453).
}

// CoinbaseChargeResponseData represents the core data returned after creating a Coinbase charge.
// It contains the necessary information to execute the on-chain payment transaction.
type CoinbaseChargeResponseData struct {
	ID string `json:"id"` // The unique ID for the Coinbase Commerce charge.

	// Addresses might be provided for specific tokens, but often omitted when paying
	// with native currency using the provided calldata/web3_data.
	Addresses map[string]string `json:"addresses,omitempty"`

	// The OpenAPI spec example showed a simple 'calldata', while the Crypto API docs
	// showed a nested 'web3_data'. We include 'web3_data' as it's more detailed
	// and likely the primary way to get transaction parameters.
	Calldata map[string]interface{} `json:"calldata,omitempty"`  // Simplified catch-all based on spec example.
	Web3Data *Web3TransferData      `json:"web3_data,omitempty"` // Detailed structure based on Crypto API docs example.

	ChainID int    `json:"chain_id"` // The chain ID specified in the request.
	Sender  string `json:"sender"`   // The sender address specified in the request.

	// Additional fields observed in Crypto API docs example (consider adding if needed):
	// CreatedAt time.Time `json:"created_at,omitempty"`
	// ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Web3TransferData mirrors the structure described in the Crypto API docs,
// containing the necessary intent data for the on-chain transaction.
type Web3TransferData struct {
	TransferIntent *TransferIntentData `json:"transfer_intent,omitempty"`
}

// TransferIntentData contains metadata and the specific calldata for the Coinbase transaction.
type TransferIntentData struct {
	Metadata *TransferMetadata `json:"metadata,omitempty"`
	CallData *TransferCallData `json:"call_data,omitempty"` // Renamed from spec 'calldata' based on docs example key
}

// TransferMetadata contains chain and contract address information.
type TransferMetadata struct {
	ChainID         int    `json:"chain_id,omitempty"`         // Chain ID for the transaction.
	ContractAddress string `json:"contract_address,omitempty"` // Target smart contract address for payment protocol.
	Sender          string `json:"sender,omitempty"`           // Sender address (may be redundant).
}

// TransferCallData contains the specific parameters needed for the payment smart contract call.
// Field types are kept as string based on the documentation example to preserve precision
// for amounts and provide hex strings for bytes/signatures.
// These correspond to the arguments expected by functions like `swapAndTransferUniswapV3Native`.
type TransferCallData struct {
	RecipientAmount   string `json:"recipient_amount,omitempty"`   // Amount recipient receives (uint256 as string).
	Deadline          string `json:"deadline,omitempty"`           // Transaction deadline (unix timestamp as string).
	Recipient         string `json:"recipient,omitempty"`          // Recipient address (payable address hex string).
	RecipientCurrency string `json:"recipient_currency,omitempty"` // Currency address (hex string, e.g., token address or zero address for native).
	RefundDestination string `json:"refund_destination,omitempty"` // Address for refunds (hex string).
	FeeAmount         string `json:"fee_amount,omitempty"`         // Fee amount (uint256 as string).
	ID                string `json:"id,omitempty"`                 // Unique intent ID (bytes16 hex string).
	Operator          string `json:"operator,omitempty"`           // Operator address (hex string).
	Signature         string `json:"signature,omitempty"`          // Transaction signature (bytes hex string).
	Prefix            string `json:"prefix,omitempty"`             // Signature prefix (bytes hex string).
}

// Note: The response structure for POST /credits/coinbase is defined in billing.go
// as:
// type CoinbaseChargeResponse struct {
//	 Data CoinbaseChargeResponseData `json:"data"`
// }
