package rbac

const (
	Page_Required     = "页码不能为空"
	PageSize_Required = "每页条数不能为空"

	Username_Required = "用户名不能为空"
	Password_Required = "密码不能为空"
	RoleId_Required   = "角色不能为空"
	Id_Required       = "ID不能为空"

	Title_Required = "标题不能为空"
	Type_Required  = "类型不能为空"
)

// #########   管理员验证器    ##############

var AdminListValidator = map[string]string{
	"Page.required":     Page_Required,
	"PageSize.required": PageSize_Required,
}

var AdminAddValidator = map[string]string{
	"Username.required": Username_Required,
	"Password.required": Password_Required,
	"RoleId.required":   RoleId_Required,
}

var AdminEditValidator = map[string]string{
	"Username.required": Username_Required,
	"RoleId.required":   RoleId_Required,
}

var AdminDeleteValidator = map[string]string{
	"Id.required": Id_Required,
}

var AdminKickLoginValidator = map[string]string{
	"AdminId.required": Id_Required,
}

var AdminLogListValidator = map[string]string{
	"Page.required":     Page_Required,
	"PageSize.required": PageSize_Required,
}

// #########   菜单验证器    ##############
var MenuAddValidator = map[string]string{
	"Title.required": Page_Required,
	"Type.required":  PageSize_Required,
}
