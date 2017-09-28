package endly_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/viant/endly"
	"github.com/viant/toolbox"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func getServiceWithWorkflow(paths ...string) (endly.Manager, endly.Service, error) {
	manager := endly.NewManager()
	service, err := manager.Service(endly.WorkflowServiceId)

	if err == nil {

		for _, workflowPath := range paths {
			context := manager.NewContext(toolbox.NewContext())
			response := service.Run(context, &endly.WorkflowLoadRequest{
				Source: endly.NewFileResource(workflowPath),
			})
			if response.Error != "" {
				return nil, nil, errors.New(response.Error)
			}
		}
	}
	return manager, service, err
}

func TestRunWorfklow(t *testing.T) {

	go StartTestServer("8765")
	time.Sleep(500 * time.Millisecond)
	manager, service, err := getServiceWithWorkflow("test/workflow/simple.csv", "test/workflow/simple_call.csv")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, manager)
	assert.NotNil(t, service)

	{
		context := manager.NewContext(toolbox.NewContext())
		response := service.Run(context, &endly.WorkflowRunRequest{
			Name: "simple",
			Params: map[string]interface{}{
				"port": "8765",
			},
		})
		assert.Equal(t, "", response.Error)
		serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
		assert.True(t, ok)
		assert.NotNil(t, serviceResponse)
	}

	{
		context := manager.NewContext(toolbox.NewContext())
		response := service.Run(context, &endly.WorkflowRunRequest{
			Name: "simple_call",
			Params: map[string]interface{}{
				"port": "8765",
			},
		})
		assert.Equal(t, "", response.Error)
		serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
		assert.True(t, ok)
		assert.NotNil(t, serviceResponse)
	}
}

func TestRunWorfklowMysql(t *testing.T) {

	manager, service, err := getServiceWithWorkflow("workflow/dockerized_mysql.csv")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, manager)
	assert.NotNil(t, service)

	{
		context := manager.NewContext(toolbox.NewContext())
		response := service.Run(context, &endly.WorkflowRunRequest{
			Name: "dockerized_mysql",
			Params: map[string]interface{}{
				"url":        "scp://127.0.0.1/",
				"credential": path.Join(os.Getenv("HOME"), "/secret/scp.json"),
			},
			Tasks: map[string]string{
				"system_stop_mysql":   "0,1",
				"system_start_docker": "0",
			},
		})
		if assert.Equal(t, "", response.Error) {
			serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
			assert.True(t, ok)
			assert.NotNil(t, serviceResponse)

			assert.Equal(t, "system_stop_mysql", serviceResponse.TasksActivities[0].Task)
			assert.Equal(t, "Does not match run criteria: $params.stopSystemMysql:true", serviceResponse.TasksActivities[0].Skipped)

			if len(serviceResponse.TasksActivities[0].ServiceActivities) > 0 {
				assert.Equal(t, "status", serviceResponse.TasksActivities[0].ServiceActivities[0].Action)
				assert.Equal(t, "", serviceResponse.TasksActivities[0].ServiceActivities[0].Skipped)

				assert.Equal(t, "stop", serviceResponse.TasksActivities[0].ServiceActivities[1].Action)
				assert.Equal(t, "", serviceResponse.TasksActivities[0].ServiceActivities[1].Skipped)
			}
		}
	}

	credential := path.Join(os.Getenv("HOME"), "secret/mysql.json")
	if toolbox.FileExists(credential) {
		context := manager.NewContext(toolbox.NewContext())
		response := service.Run(context, &endly.WorkflowRunRequest{
			Name: "dockerized_mysql",
			Params: map[string]interface{}{
				"url":                 "scp://127.0.0.1/",
				"credential":          path.Join(os.Getenv("HOME"), "/secret/scp.json"),
				"stopSystemMysql":     true,
				"configUrl":           endly.NewFileResource("test/docker/my.cnf").URL,
				"confifUrlCredential": "",
				"serviceInstanceName": "dockerizedMysql1",
			},
		})

		if assert.Equal(t, "", response.Error) {
			serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
			assert.True(t, ok)
			assert.NotNil(t, serviceResponse)

			assert.Equal(t, "system_stop_mysql", serviceResponse.TasksActivities[0].Task)
			assert.Equal(t, "", serviceResponse.TasksActivities[0].Skipped)

			assert.Equal(t, "status", serviceResponse.TasksActivities[0].ServiceActivities[0].Action)
			assert.Equal(t, "", serviceResponse.TasksActivities[0].ServiceActivities[0].Skipped)
			assert.Equal(t, "stop", serviceResponse.TasksActivities[0].ServiceActivities[1].Action)

			assert.Equal(t, "status", serviceResponse.TasksActivities[1].ServiceActivities[0].Action)
			assert.Equal(t, "start", serviceResponse.TasksActivities[1].ServiceActivities[1].Action)
		}

	}

}

func TestRunWorfklowAerospike(t *testing.T) {

	manager, service, err := getServiceWithWorkflow("workflow/dockerized_aerospike.csv")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, manager)
	assert.NotNil(t, service)
	credential := path.Join(os.Getenv("HOME"), "secret/scp.json")
	if toolbox.FileExists(credential) {
		context := manager.NewContext(toolbox.NewContext())
		response := service.Run(context, &endly.WorkflowRunRequest{
			Name: "dockerized_aerospike",
			Params: map[string]interface{}{
				"url":                 "scp://127.0.0.1/",
				"credential":          path.Join(os.Getenv("HOME"), "/secret/scp.json"),
				"configUrl":           endly.NewFileResource("test/workflow/aerospike.conf").URL,
				"confifUrlCredential": "",
				"serviceInstanceName": "dockerizedAerospike1",
			},
		})
		if assert.Equal(t, "", response.Error) {
			serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
			assert.True(t, ok)
			assert.NotNil(t, serviceResponse)
			assert.Equal(t, "system_start_docker", serviceResponse.TasksActivities[0].Task)
			assert.Equal(t, "prepare_config", serviceResponse.TasksActivities[1].Task)
			assert.Equal(t, "docker_run_aerospike", serviceResponse.TasksActivities[2].Task)
		}

	}

}

func TestRunWorfkloVCMavenwBuild(t *testing.T) {

	manager, service, err := getServiceWithWorkflow("workflow/vc_maven_build.csv")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, manager)
	assert.NotNil(t, service)
	credential := path.Join(os.Getenv("HOME"), "secret/scp.json")
	if toolbox.FileExists(credential) {
		baseSvnUrlFile := path.Join(os.Getenv("HOME"), "baseSvnUrl")
		if toolbox.FileExists(baseSvnUrlFile) {
			baseSvnUrl, err := endly.NewFileResource(path.Join(os.Getenv("HOME"), "baseSvnUrl")).DownloadText()
			baseSvnUrl = strings.Trim(baseSvnUrl, " \r\n")
			assert.Nil(t, err)
			context := manager.NewContext(toolbox.NewContext())
			response := service.Run(context, &endly.WorkflowRunRequest{
				Name: "vc_maven_build",
				Params: map[string]interface{}{
					"jdkVersion":       "1.7",
					"originUrl":        baseSvnUrl + "/common",
					"originCredential": path.Join(os.Getenv("HOME"), "/secret/svn_ci.json"),
					"originType":       "svn",
					"targetUrl":        "file:///tmp/ci_common",
					"targetCredential": "",
					"buildGoal":        "install",
					"buildArgs":        "-Dmvn.test.skip",
				},
			})
			if assert.Equal(t, "", response.Error) {
				serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
				assert.True(t, ok)
				assert.NotNil(t, serviceResponse)

			}
		}

	}

}

func TestRunWorfkloTomcatApp(t *testing.T) {

	manager, service, err := getServiceWithWorkflow("workflow/tomcat.csv")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, manager)
	assert.NotNil(t, service)
	credential := path.Join(os.Getenv("HOME"), "secret/scp.json")
	if toolbox.FileExists(credential) {
		baseSvnUrlFile := path.Join(os.Getenv("HOME"), "baseSvnUrl")
		if toolbox.FileExists(baseSvnUrlFile) {
			baseSvnUrl, err := endly.NewFileResource(path.Join(os.Getenv("HOME"), "baseSvnUrl")).DownloadText()
			baseSvnUrl = strings.Trim(baseSvnUrl, " \r\n")
			assert.Nil(t, err)
			context := manager.NewContext(toolbox.NewContext())

			{
				response := service.Run(context, &endly.WorkflowRunRequest{
					Name: "tomcat",
					Params: map[string]interface{}{
						"jdkVersion":         "1.7",
						"targetHost":         "127.0.0.1",
						"targetCredential":   path.Join(os.Getenv("HOME"), "/secret/scp.json"),
						"appDirectory":       "/tmp/app1",
						"tomcatVersion":      "7.0.81",
						"tomcatMajorVersion": "7",
					},
					Task: "install",
				})
				if assert.Equal(t, "", response.Error) {
					serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
					assert.True(t, ok)
					assert.NotNil(t, serviceResponse)

				}
			}
			{
				response := service.Run(context, &endly.WorkflowRunRequest{
					Name: "tomcat",
					Params: map[string]interface{}{
						"jdkVersion":       "1.7",
						"targetHost":       "127.0.0.1",
						"targetCredential": path.Join(os.Getenv("HOME"), "/secret/scp.json"),
						"appDirectory":     "/tmp/app1",
						"catalinaOpts":     "-Xms2g -Xmx6g -XX:MaxPermSize=512m",
					},
					Task: "start",
				})
				if assert.Equal(t, "", response.Error) {
					serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
					assert.True(t, ok)
					assert.NotNil(t, serviceResponse)

				}
			}

			time.Sleep(2 * time.Second)
			{
				response := service.Run(context, &endly.WorkflowRunRequest{
					Name: "tomcat",
					Params: map[string]interface{}{
						"jdkVersion":       "1.7",
						"targetHost":       "127.0.0.1",
						"targetCredential": path.Join(os.Getenv("HOME"), "/secret/scp.json"),
						"appDirectory":     "/tmp/app1",
						"catalinaOpts":     "-Xms2g -Xmx6g -XX:MaxPermSize=512m",
					},
					Task: "stop",
				})
				if assert.Equal(t, "", response.Error) {
					serviceResponse, ok := response.Response.(*endly.WorkflowRunResponse)
					assert.True(t, ok)
					assert.NotNil(t, serviceResponse)

				}
			}
		}

	}

}