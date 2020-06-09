/**------------------------------------------------------------**
 * @filename project/xxx.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/4 18:14
 * @desc     go.jd100.com - main - summary
 **------------------------------------------------------------**/
package project

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	_tplTypeMain = iota
	_tplTypeConf
	_tplTypeService
	_tplTypeModel
	_tplTypeDao
	_tplTypeDaoMysql
	_tplTypeHTTPServer
	_tplTypeAPIProto
	_tplTypeGRPCServer

	_tplTypeServiceTest

	_tplTypeChangeLog
	_tplTypeContributors
	_tplTypeReadme
	_tplTypeAppToml
	_tplTypeGomod

	_tplTypeAPIGogen

	_tplDevOpsDockerfile
	_tplDevOpsDeploySh
	_tplDevOpsJenkinsfile
	_tplDevOpsJenkinsfileA
	_tplDevOpsJenkinsfileD
	_tplDevOpsJenkinsfileS
	_tplDevOpsJenkinsfileT
	_tplDevOpsMakefile

	_tplDevOpsHelmChart
	_tplDevOpsHelmConfigMap
	_tplDevOpsHelmValues
	_tplDevOpsTemplateD
	_tplDevOpsTemplateI
	_tplDevOpsTemplateS
)

const (
	DocType int = iota + 1
	CodeType
	ConfType
	DeployType
	GRPCType
)

var (
	P           Project
	genFileList = map[int]map[int]string{
		DocType: {
			_tplTypeChangeLog:    "/CHANGELOG.md",
			_tplTypeContributors: "/CONTRIBUTORS.md",
			_tplTypeReadme:       "/README.md",
		},
		CodeType: {
			_tplTypeMain:       "/cmd/main.go",
			_tplTypeConf:       "/conf/conf.go",
			_tplTypeDao:        "/dao/dao.go",
			_tplTypeDaoMysql:   "/dao/mysql.go",
			_tplTypeHTTPServer: "/server/http/server.go",
			_tplTypeService:    "/service/service.go",
			_tplTypeModel:      "/model/model.go",
		},
		ConfType: {
			_tplTypeGomod: "/go.mod",
			// init config
			_tplTypeAppToml: "/bin/app.toml",
		},
		DeployType: {
			_tplDevOpsDockerfile:   "/install/build/Dockerfile",
			_tplDevOpsDeploySh:     "/install/build/deploy_docker.sh",
			_tplDevOpsJenkinsfile:  "/install/build/Jenkinsfile",
			_tplDevOpsJenkinsfileA: "/install/build/Jenkinsfile-ali",
			_tplDevOpsJenkinsfileD: "/install/build/Jenkinsfile-dev",
			_tplDevOpsJenkinsfileS: "/install/build/Jenkinsfile-sonar",
			_tplDevOpsJenkinsfileT: "/install/build/Jenkinsfile-test",
			_tplDevOpsMakefile:     "/install/build/Makefile",

			_tplDevOpsHelmChart:     "/install/helm/{pn}/Chart.yaml",
			_tplDevOpsHelmConfigMap: "/install/helm/{pn}/configmap-{pn}.yaml",
			_tplDevOpsHelmValues:    "/install/helm/{pn}/values.yaml",
			//_tplDevOpsTemplateD:     "/install/helm/{pn}/templates/deployment.yaml",
			//_tplDevOpsTemplateI:     "/install/helm/{pn}/templates/ingress.yaml",
			//_tplDevOpsTemplateS:     "/install/helm/{pn}/templates/service.yaml",
		},
		GRPCType: {
			_tplTypeGRPCServer: "/server/grpc/server.go",
			_tplTypeAPIProto:   "/api/api.proto",
			_tplTypeAPIGogen:   "/api/generate.go",
		},
	}
	// tpls type => content
	tpls = map[int]string{
		_tplTypeDao:          _tplDao,
		_tplTypeDaoMysql:     _tplDaoMysql,
		_tplTypeHTTPServer:   _tplHTTPServer,
		_tplTypeAPIProto:     _tplAPIProto,
		_tplTypeMain:         _tplMain,
		_tplTypeConf:         _tplConf,
		_tplTypeService:      _tplService,
		_tplTypeChangeLog:    _tplChangeLog,
		_tplTypeContributors: _tplContributors,
		_tplTypeReadme:       _tplReadme,
		_tplTypeAppToml:      _tplAppToml,
		_tplTypeModel:        _tplModel,
		_tplTypeGomod:        _tplGoMod,
		_tplTypeAPIGogen:     _tplGogen,

		_tplTypeServiceTest: _tplServiceTest,

		_tplDevOpsDockerfile:   _tplDODockerfile,
		_tplDevOpsDeploySh:     _tplDODeploySh,
		_tplDevOpsJenkinsfile:  _tplDOJenkinsfile,
		_tplDevOpsJenkinsfileA: _tplDOJenkinsfileA,
		_tplDevOpsJenkinsfileD: _tplDOJenkinsfileD,
		_tplDevOpsJenkinsfileS: _tplDOJenkinsfileS,
		_tplDevOpsJenkinsfileT: _tplDOJenkinsfileT,
		_tplDevOpsMakefile:     _tplDOMakefile,

		_tplDevOpsHelmChart:     _tplDOHelmChart,
		_tplDevOpsHelmConfigMap: _tplDOHelmConfigMap,
		_tplDevOpsHelmValues:    _tplDOHelmValues,
		_tplDevOpsTemplateD:     _tplDOTemplateD,
		_tplDevOpsTemplateI:     _tplDOTemplateI,
		_tplDevOpsTemplateS:     _tplDOTemplateS,
	}
)

// project project config
type Project struct {
	Name      string
	Owner     string
	Path      string
	WithGRPC  bool
	Here      bool
	Module    string // 支持项目的自定义module名 （go.mod init）
	Date      string
	Namespace string
}

func Create() (err error) {
	if P.WithGRPC {
		tpls[_tplTypeHTTPServer] = _tplPBHTTPServer
		tpls[_tplTypeGRPCServer] = _tplGRPCServer
		tpls[_tplTypeService] = _tplGPRCService
		tpls[_tplTypeMain] = _tplGRPCMain
	} else {
		delete(genFileList[GRPCType], _tplTypeGRPCServer)
	}
	if err = os.MkdirAll(P.Path, 0755); err != nil {
		return
	}
	for ty, files := range genFileList {
		for t, v := range files {
			v = strings.Replace(v, "{pn}", P.Name, -1)
			i := strings.LastIndex(v, "/")
			if i > 0 {
				dir := v[:i]
				if err = os.MkdirAll(P.Path+dir, 0755); err != nil {
					return
				}
			}
			if err = write(P.Path+v, tpls[t]); err != nil {
				return
			}

			if ty == CodeType && !strings.Contains(v, "main.go") && !strings.Contains(v, "mysql.go") {
				pathL := strings.Split(v, "/")
				indx := len(pathL) - 2
				P.Namespace = pathL[indx] + "_test"
				var testName = strings.Replace(v, ".go", "_test.go", 1)
				if err = write(P.Path+testName, tpls[_tplTypeServiceTest]); err != nil {
					return
				}
			}
		}
	}

	if P.WithGRPC {
		if err = genpb(); err != nil {
			return
		}
	}
	return
}

func genpb() error {
	cmd := exec.Command("jiny", "tool", "protoc", P.Name+"/api/api.proto")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func write(name, tpl string) (err error) {
	data, err := parse(tpl)
	if err != nil {
		return
	}
	return ioutil.WriteFile(name, data, 0644)
}

func parse(s string) ([]byte, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, P); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
