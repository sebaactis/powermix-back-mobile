package coffeeji

import "encoding/json"

type OrderListResponse struct {
	Code    int           `json:"code"`
	Success bool          `json:"success"`
	Data    OrderListData `json:"data"`
	Msg     string        `json:"msg"`
}

type OrderListData struct {
	Records []OrderRecord `json:"records"`
}

type OrderRecord struct {
	GoodsName string `json:"goodsName"`
}

type VoucherResponse struct {
	Code    int             `json:"code"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Msg     string          `json:"msg"`
}
