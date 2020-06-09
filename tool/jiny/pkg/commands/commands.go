/**------------------------------------------------------------**
 * @filename commands/xxx.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/13 15:27
 * @desc     go.jd100.com - commands - summary
 **------------------------------------------------------------**/
package commands

import "github.com/spf13/cobra"

func AddCommands(cmd *cobra.Command) {
	addVersion(cmd)
	addCreate(cmd)
}
