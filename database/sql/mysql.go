/**------------------------------------------------------------**
 * @filename sql/mysql.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/10/15 09:41
 * @desc     go.jd100.com - sql - mysql adapter db
 **------------------------------------------------------------**/
package sql

func NewMySQL(c *Config) (db *DB) {
	return newDB(_mysql, c)
}
