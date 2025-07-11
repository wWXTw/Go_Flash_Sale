package e

// 定义返回消息
var MsgFlags = map[int]string{
	SUCCESS:            "成功!",
	NotExistIdentifier: "该第三方账号未绑定", // ???
	ERROR:              "致命错误!",
	InvalidParams:      "请求参数有误!",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
