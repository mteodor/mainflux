package api

type addThingReq struct {
	token       string
	ExternalID  string `json:"externalid"`
	ExternalKey string `json:"externalkey"`
}

func (req addThingReq) validate() error {
	if req.ExternalID == "" || req.ExternalKey == "" {
		return errUnauthorized
	}
	return nil
}
