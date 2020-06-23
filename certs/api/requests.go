package api

type addCertsReq struct {
	token      string
	ThingID    string `json:"thing_id"`
	RsaBits    int    `json:"rsa_bits"`
	Encryption string `json:"encryption"`
	Valid      string `json:"valid"`
}

func (req addCertsReq) validate() error {
	if req.ThingID == "" {
		return errUnauthorized
	}
	return nil
}
