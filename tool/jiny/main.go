package main

import (
	"log"

	"go.jd100.com/medusa/tool/jiny/pkg/commands"

	"github.com/spf13/cobra"
)

/**--------------------------------------**
 * 程序入口
 **--------------------------------------**/
func main() {
	cmds := &cobra.Command{
		Use:   "jiny",
		Short: "快速创建基于Medusa框架的Golang项目，及部署配置",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	commands.AddCommands(cmds)

	if err := cmds.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
