package nobitex

type (
	PositionsResponse struct {
		Status    string     `json:"status"`
		Positions []Position `json:"positions"`
	}

	Position struct {
		ID                   int     `json:"id"`
		CreatedAt            string  `json:"createdAt"`
		SrcCurrency          string  `json:"srcCurrency"`
		DstCurrency          string  `json:"dstCurrency"`
		Side                 string  `json:"side"`
		Status               string  `json:"status"`
		MarginType           string  `json:"marginType"`
		Collateral           string  `json:"collateral"`
		Leverage             string  `json:"leverage"`
		OpenedAt             string  `json:"openedAt"`
		ClosedAt             *string `json:"closedAt"`
		LiquidationPrice     string  `json:"liquidationPrice"`
		EntryPrice           string  `json:"entryPrice"`
		ExitPrice            *string `json:"exitPrice"`
		DelegatedAmount      string  `json:"delegatedAmount"`
		Liability            string  `json:"liability"`
		TotalAsset           string  `json:"totalAsset"`
		MarginRatio          string  `json:"marginRatio"`
		LiabilityInOrder     string  `json:"liabilityInOrder"`
		AssetInOrder         string  `json:"assetInOrder"`
		UnrealizedPNL        string  `json:"unrealizedPNL"`
		UnrealizedPNLPercent string  `json:"unrealizedPNLPercent"`
		ExpirationDate       string  `json:"expirationDate"`
		ExtensionFee         string  `json:"extensionFee"`
		MarkPrice            string  `json:"markPrice"`
	}
)

type OrderResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Order   struct {
		ID int `json:"id"`
	} `json:"order"`
}

type CancelOrderResponse struct {
	Status        string `json:"status"`
	Code          string `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
	UpdatedStatus string `json:"updatedStatus"`
}

type BalanceResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Wallets map[string]struct {
		ID      int    `json:"id"`
		Balance string `json:"balance"`
		Blocked string `json:"blocked"`
	} `json:"wallets"`
}
