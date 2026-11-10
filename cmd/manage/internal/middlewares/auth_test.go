package middlewares

import (
	"testing"

	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
)

func TestAPIPermissionMatchesRegisteredRoute(t *testing.T) {
	endpoint := types.BusinessEndpoint{Method: "GET", Path: "/stats/team"}
	node := model.Menu{
		Id: 1, Type: model.MENU_TYPE_NODE, Status: model.MENU_STATUS_ON,
		ApiPath: endpoint.Path, ApiMethod: endpoint.Method,
		ApiPermission: model.PermissionCode(endpoint.Permission()),
	}
	if !apiPermissionAllowed(endpoint, []int{node.Id}, []model.Menu{node}) {
		t.Fatal("matching enabled permission was rejected")
	}
	alias := types.BusinessEndpoint{Method: "GET", Path: "/stats.team"}
	if alias.Permission() != endpoint.Permission() {
		t.Fatal("test must use routes with identical derived permission codes")
	}
	if apiPermissionAllowed(alias, []int{node.Id}, []model.Menu{node}) {
		t.Fatal("permission for another route with the same derived code was accepted")
	}
	node.Status = model.MENU_STATUS_OFF
	if apiPermissionAllowed(endpoint, []int{node.Id}, []model.Menu{node}) {
		t.Fatal("disabled permission was accepted")
	}
}
