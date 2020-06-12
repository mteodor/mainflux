package api

type addCertsReq struct {
	token   string
	ThingID string `json:"thing_id"`
}

func (req addCertsReq) validate() error {
	if req.ThingID == "" {
		return errUnauthorized
	}
	return nil
}
