package types

import "example.com/go-service-kit/shared/types/page"

type IpWhitelistListRequest struct {
	IP string `form:"ip"`
	page.PageRequest
}

type IpWhitelistAddRequest struct {
	IP     string `json:"ip" binding:"required"`
	Remark string `json:"remark"`
}

type IpWhitelistDeleteRequest struct {
	Ids string `form:"ids"  binding:"required"`
}

type IpWhitelistStatusRequest struct {
	Status *int `json:"status" binding:"required,oneof=0 1"`
}
