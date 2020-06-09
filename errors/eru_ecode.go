/**------------------------------------------------------------**
 * @filename errors/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-09-23 15:32
 * @desc     dean - errors -
 **------------------------------------------------------------**/
package errors

var (
	DeanSubjectNoExist      = add(70100) // 科目不存在


	DeanPhaseErr            = add(70110) // 学段错误
	DeanPhaseSubjectsNoData = add(70120) // 该学段暂未分配学科

	DeanEruOrderNoDuplicate = add(70140) // 订单NO已存在
	DeanEruMemberNoExist    = add(70150) // 订单学员不存在
	DeanEruGuiderNoExist    = add(70200) // 学管师不存在

	DeanEruGuiderMProxyOErr = add(71000) // 教务学管师主代理1账号错误
	DeanEruGuiderMProxyTErr = add(71001) // 教务学管师主代理2账号错误

	// SubjectNoExist = ""
)