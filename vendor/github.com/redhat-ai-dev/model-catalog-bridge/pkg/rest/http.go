package rest

type PostBody struct {
	Body                     []byte `json:"body"`
	LastUpdateTimeSinceEpoch string `json:"lastUpdateTimeSinceEpoch"`
	ModelCardKey             string `json:"modelCardKey"`
	ModelCard                string `json:"modelCard"`
}
