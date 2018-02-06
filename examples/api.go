package main

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func createAPI() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	sb := c.NewAPIBuilder()
	// 必选项
	sb.Name("用户API")
	// 设置URL规则，匹配所有开头为/api/user的请求
	sb.MatchURLPattern("/api/user/(.+)")
	// 匹配GET请求
	sb.MatchMethod("GET")
	// 匹配所有请求
	sb.MatchMethod("*")
	// 不启动
	sb.Down()
	// 启用
	sb.UP()
	// 分发到Cluster 1
	sb.AddDispatchNode(1)

	// 可选项
	// 匹配所有host
	sb.MatchDomain("user.xxx.com")
	// 增加访问黑名单
	sb.AddBlacklist("192.168.0.1", "192.168.1.*", "192.168.*")
	// 增加访问报名单
	sb.AddWhitelist("192.168.3.1", "192.168.3.*", "192.168.*")
	// 移除黑白名单
	sb.RemoveBlacklist("192.168.0.1") // 剩余："192.168.1.*", "192.168.*"
	sb.RemoveWhitelist("192.168.3.1") // 剩余："192.168.3.*", "192.168.*"

	// 增加默认值
	sb.DefaultValue([]byte("{\"value\", \"default\"}"))
	// 为默认值增加header
	sb.AddDefaultValueHeader("token", "xxxxx")
	// 为默认值增加Cookie
	sb.AddDefaultValueCookie("sid", "xxxxx")

	// 设置鉴权filter，那么名为jwt的插件就会拦截这个请求，检查并解析jwt的token
	sb.AuthPlugin("jwt")

	// 设置这个API访问需要的权限，同时满足perm1和perm2的用户才可以访问这个API，需要配合业务自己的权限插件
	sb.AddPerm("PERM1")
	sb.AddPerm("PERM2")

	// 给分发到cluster 1 的节点增加校验
	// 必须包含name的query string param，并且必须是字母
	param := metapb.Parameter{
		Name:   "name",
		Source: metapb.QueryString,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须json body的json必须包含name属性，并且必须是字母
	// 可以是级联属性，必须 user.name，那么就表示json body的json中必须包含 {"user": {"name": "xxxx"}}
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.JSONBody,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须包含name的cookie param，并且必须是字母
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.Cookie,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须包含name的form data，并且必须是字母
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.FormData,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 增加一个转发，
	sb.AddDispatchNode(2)
	//  重写转发到1的URL
	sb.DispatchNodeURLRewrite(1, "/api/user/base/$1")
	//  重写转发到2的URL
	sb.DispatchNodeURLRewrite(2, "/api/user/account/$1")
	// 设置转发到1的返回值的属性为 base
	sb.DispatchNodeValueAttrName(1, "base")
	// 设置转发到1的返回值的属性为 account
	sb.DispatchNodeValueAttrName(2, "account")
	// 经过上面的设置，gateway聚合的返回值为：{"base": {1 返回的json}, "account": {2 返回的JSON}}

	// 重新定义渲染结果，转为：{"base": {"feild1": xx, "feild2": xx}, "account": {"feild1": xx}}
	sb.AddRenderObject("base", "feild1", "base.user.feild1", "field2", "base.user.field2")
	sb.AddRenderObject("account", "feild1", "account.field1")

	// 清空
	sb.NoRenderTemplate()

	// 重新定义渲染结果，转为：{"obj1": {"felid1": xxx, "filed2": xxx}, "account_field1": "xxx"}
	sb.AddRenderObject("obj1", "felid1", "base.user.felid1", "felid2", "base.user.felid2")
	sb.AddFlatRenderObject("account_field1", "account.felid1")

	id, err := sb.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("api id is: %d", id)
	return nil
}

func updateAPI(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	api, err := c.GetAPI(id)
	if err != nil {
		return err
	}

	// 下线API
	c.NewAPIBuilder().Use(*api).UP().Commit()
	// 上线API
	c.NewAPIBuilder().Use(*api).UP().Commit()

	// 修改你期望修改的字段
	sb := c.NewAPIBuilder().Use(*api)
	// 匹配所有host
	sb.MatchDomain("user.xxx.com")
	// 增加访问黑名单
	sb.AddBlacklist("192.168.0.1", "192.168.1.*", "192.168.*")
	// 增加访问报名单
	sb.AddWhitelist("192.168.3.1", "192.168.3.*", "192.168.*")
	// 移除黑白名单
	sb.RemoveBlacklist("192.168.0.1") // 剩余："192.168.1.*", "192.168.*"
	sb.RemoveWhitelist("192.168.3.1") // 剩余："192.168.3.*", "192.168.*"

	// 增加默认值
	sb.DefaultValue([]byte("{\"value\", \"default\"}"))
	// 为默认值增加header
	sb.AddDefaultValueHeader("token", "xxxxx")
	// 为默认值增加Cookie
	sb.AddDefaultValueCookie("sid", "xxxxx")

	// 设置鉴权filter，那么名为jwt的插件就会拦截这个请求，检查并解析jwt的token
	sb.AuthPlugin("jwt")

	// 设置这个API访问需要的权限，同时满足perm1和perm2的用户才可以访问这个API，需要配合业务自己的权限插件
	sb.AddPerm("PERM1")
	sb.AddPerm("PERM2")

	// 给分发到cluster 1 的节点增加校验
	// 必须包含name的query string param，并且必须是字母
	param := metapb.Parameter{
		Name:   "name",
		Source: metapb.QueryString,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须json body的json必须包含name属性，并且必须是字母
	// 可以是级联属性，必须 user.name，那么就表示json body的json中必须包含 {"user": {"name": "xxxx"}}
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.JSONBody,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须包含name的cookie param，并且必须是字母
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.Cookie,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 必须包含name的form data，并且必须是字母
	param = metapb.Parameter{
		Name:   "name",
		Source: metapb.FormData,
	}
	sb.AddDispatchNodeValidation(1, param, "[a-zA-Z]+", true)

	// 增加一个转发，
	sb.AddDispatchNode(2)
	//  重写转发到1的URL
	sb.DispatchNodeURLRewrite(1, "/api/user/base/$1")
	//  重写转发到2的URL
	sb.DispatchNodeURLRewrite(2, "/api/user/account/$1")
	// 设置转发到1的返回值的属性为 base
	sb.DispatchNodeValueAttrName(1, "base")
	// 设置转发到1的返回值的属性为 account
	sb.DispatchNodeValueAttrName(2, "account")
	// 经过上面的设置，gateway聚合的返回值为：{"base": {1 返回的json}, "account": {2 返回的JSON}}
	// 完成修改
	sb.Commit()

	return nil
}

func deleteAPI(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	return c.RemoveAPI(id)
}
