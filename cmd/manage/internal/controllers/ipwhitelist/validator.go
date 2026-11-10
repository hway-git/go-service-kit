package ipwhitelist

const (
	IP_Required = "IP不能为空"
	ID_Required = "ID不能为空"
)

var (
	IpWhitelistListValidator = map[string]string{}

	IpWhitelistAddValidator = map[string]string{
		"IP.required": IP_Required,
	}

	IpWhitelistDeleteValidator = map[string]string{
		"ID.required": ID_Required,
	}

	IpWhitelistStatusValidator = map[string]string{}
)
