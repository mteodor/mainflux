package api

type addThingReq struct {
	token       string
	Name        string `json:"name"`
	ExternalID  string `json:"external_id"`
	ExternalKey string `json:"external_key"`
}

type addCertReq struct {
	token      string
	ThingID    string `json:"thing_id"`
	RsaBits    int    `json:"rsa_bits"`
	HoursValid string `json:"days_valid"`
}

func (req addThingReq) validate() error {
	if req.ExternalID == "" || req.ExternalKey == "" {
		return errUnauthorized
	}
	return nil
}

func (req addCertReq) validate() error {
	if req.ThingID == "" || req.RsaBits < 0 || req.HoursValid == "" {
		return errMalformedEntity
	}
	return nil
}
