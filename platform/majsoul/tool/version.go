package tool

type majsoulVersion struct {
	Code       string `json:"code"`    // 설명
	ResVersion string `json:"version"` // 설명
}

func GetMajsoulVersion(apiGetVersionURL string) (version *majsoulVersion, err error) {
	version = &majsoulVersion{}
	err = get(apiGetVersionURL, version)
	return
}
