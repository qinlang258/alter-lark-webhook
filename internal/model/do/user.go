// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure of table user for DAO operations like Where/Data.
type User struct {
	g.Meta        `orm:"table:user, do:true"`
	Id            interface{} // 用户id
	KeycloakId    interface{} // keycloak_id
	OpenId        interface{} // lark_id
	Email         interface{} // 企业邮箱地址
	Mobile        interface{} // 手机号码
	Name          interface{} // 姓名
	DepartmentIds interface{} // 部门id,一个人可能属于多个部门
	UpdateTime    *gtime.Time // 更新时间
	LastLoginTime *gtime.Time // 最近一次登录时间, 后期可以清理离职用户, 否则这张表只增加不减少
	Roles         interface{} // 用户角色信息,数组形式
}
