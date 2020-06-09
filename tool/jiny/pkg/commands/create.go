/**------------------------------------------------------------**
 * @filename commands/xxx.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/13 15:32
 * @desc     go.jd100.com - commands - summary
 **------------------------------------------------------------**/
package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"go.jd100.com/medusa/tool/jiny/project"

	"github.com/spf13/cobra"
)

func AddProjectCOptions(cmd *cobra.Command, p *project.Project) {
	cmd.Flags().StringVarP(&p.Module, "module", "m", project.P.Name, "project type name for project")
	cmd.Flags().StringVarP(&p.Owner, "owner", "o", "AuthorName", "project owner for create project")
	cmd.Flags().StringVarP(&p.Path, "path", "p", "", "project path for create project")
}

func init() {

}

func addCreate(cmd *cobra.Command) {
	create := &cobra.Command{
		Use:   "new PROJECT_NAME",
		Short: "创建新项目",
		Long:  `快速创建基于Medusa的Golang项目，你只需要关注业务实现就好，其他一切给你搞定！`,
		Example: `
  # 创建新项目 jiny new project_name -o owner -m module
  # project_name 最好为单个有意义的单词：
  #   jiny new member -o jinycoo -m service
  # 其中有两个参数，-o 和 -m
  # -o author   项目创建人 注册创建人姓名 -o 后最好是自己姓名全拼，有利于跟公司邮箱一致
  # -m module   项目类型模块名称
  # 依据规则可选择：： 后台 = admin  接口 = interface  服务 = service`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalf("required project name")
			}
			project.P.Name = args[0]
			project.P.Date = time.Now().Format("2006-01-02 15:04")

			if project.P.Path != "" {
				project.P.Path = path.Join(project.P.Path, project.P.Name)
			} else {
				pwd, _ := os.Getwd()
				project.P.Path = path.Join(pwd, project.P.Name)
			}
			// creata a project
			if err := project.Create(); err != nil {
				log.Fatalf("create new project err (%v)", err)
			}
			fmt.Printf("Project: %s\n", project.P.Name)
			fmt.Printf("Owner: %s\n", project.P.Owner)
			fmt.Printf("Module Name: %s\n", project.P.Module)
			fmt.Printf("WithGRPC: %t\n", project.P.WithGRPC)
			fmt.Printf("Directory: %s\n\n", project.P.Path)
			fmt.Println("The application has been created.")
		},
	}
	AddProjectCOptions(create, &project.P)
	cmd.AddCommand(create)
}
