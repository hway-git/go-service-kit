package response

const (
	CODE_SUCCESS         = 0    // 成功
	INVALID_TOKEN    int = 1001 // 无效 token
	BAN_USER         int = 1002 // 封禁用户
	NOT_ALLOWED_PAGE int = 1003 // 页面无权限
	NOT_ALLOWED_API  int = 1004 // 接口无权限

	INVALID_PARAMS int = 2001 // 参数错误
	EXIST_USER     int = 2002 // 用户（邮箱）已存在
	NOT_FOUND_USER int = 2003 // 用户（邮箱）不存在
	PASSWORD_ERROR int = 2004 // 密码错误

	SERVER_ERROR int = 3001 // 服务器内部

)

// http返回码
const (
	HTTP_API_OK int = 200 // 请求成功

	HTTP_API_ERROR        int = 901 // 请求失败
	HTTP_INVALID_TOKEN    int = 902 // 无效 token
	HTTP_NOT_ALLOWED_PAGE int = 903 // 页面无权限
	HTTP_NOT_ALLOWED_API  int = 904 // 接口无权限
	HTTP_RESTRICT_USE     int = 905 // 限制使用模式
	HTTP_BAN_USER         int = 906 // 封禁用户
	HTTP_NOT_ALLOWED_IP   int = 907 // 不在IP白名单
)
