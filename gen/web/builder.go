package web

import (
	"bytes"
	"fmt"
	"github.com/viant/endly/util"
	"github.com/viant/neatly"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"github.com/viant/toolbox/storage"
	"github.com/viant/toolbox/url"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"path"
	"strings"
)

type builder struct {
	baseURL              string
	destURL              string
	destService, storage storage.Service
	registerDb           Map
	services             Map
	tags                 []string
	createDb             Map
	dbMeta               *DbMeta
	populateDb           Map
}

func (b *builder) addDatastore(assets map[string]string, meta *DbMeta, request *Datastore) error {
	if b.createDb.Has(request.Name) {
		return nil
	}
	b.dbMeta = meta
	var state = data.NewMap()
	state.Put("db", request.Name)
	init, err := b.NewAssetMap(assets, "init.yaml", state)
	if err != nil {
		return err
	}

	if b.registerDb, err = b.NewAssetMap(assets, "register.yaml", state); err != nil {
		return err
	}

	//ddl/schema.ddl
	if meta.Schema != "" {
		var scriptURL = fmt.Sprintf("datastore/%v/schema.sql", request.Name)
		schema, ok := assets[meta.Schema]
		if !ok {
			return fmt.Errorf("unable locate %v schema : %v", request.Driver, meta.Schema)
		}
		b.UploadToEndly(scriptURL, strings.NewReader(toolbox.AsString(schema)))
		state.Put("script", scriptURL)
		script, err := b.NewMapFromURI("datastore/script.yaml", state)
		if err != nil {
			return err
		}
		init.Put("scripts", script.Get("scripts"))
	}
	b.createDb.Put(request.Name, init)

	//dictionary
	if meta.Dictionary != "" {
		dictionaryURL := fmt.Sprintf("datastore/%v/dictionary", request.Name)

		for k, v := range assets {
			if strings.HasPrefix(k, meta.Dictionary) {
				k = string(k[len(meta.Dictionary):])
				assetURL := path.Join(dictionaryURL, k)
				b.UploadToEndly(assetURL, strings.NewReader(v))
			}
		}
		state.Put("dictionary", dictionaryURL)
		prepare, err := b.NewMapFromURI("datastore/prepare.yaml", state)
		if err != nil {
			return err
		}
		b.populateDb.Put(request.Name, prepare)
	}
	return nil
}

func (b *builder) addDatastoreService(assets map[string]string, meta *DbMeta, request *Datastore) error {
	if b.services.Has(request.Driver) {
		return nil
	}
	var service = NewMap()
	service.Put("workflow", meta.Service)
	service.Put("name", "endly_"+request.Name)
	var version = request.Version
	if version == "" {
		version = meta.Version
	}
	if version == "" {
		version = "latest"
	}
	service.Put("version", version)
	if meta.Credentials != "" {
		service.Put("credentials", meta.Credentials)
	}
	if request.Config && meta.Config != "" {
		config, ok := assets[meta.Config]
		if !ok {
			return fmt.Errorf("unable locate %v service config: %v", request.Driver, meta.Config)
		}
		var configURL = fmt.Sprintf("datastore/%v", meta.Config)
		b.UploadToEndly(configURL, strings.NewReader(toolbox.AsString(config)))
		service.Put("config", configURL)
	}

	b.services.Put(request.Driver, service)
	if meta.Tag != "" {
		b.tags = append(b.tags, meta.Tag)
	}
	var state = data.NewMap()
	state.Put("db", request.Name)
	state.Put("driver", request.Driver)
	ReadIp, _ := b.NewMapFromURI("datastore/ip.yaml", state)
	b.services.Put(request.Driver+"-ip", ReadIp)
	return nil
}

func (b *builder) asMap(text string, state data.Map) (Map, error) {
	aMap := yaml.MapSlice{}
	if state != nil {
		text = state.ExpandAsText(text)
	}
	err := yaml.NewDecoder(strings.NewReader(text)).Decode(&aMap)
	if err != nil {
		err = fmt.Errorf("failed to decode %v, %v", text, err)
	}
	var result = mapSlice(aMap)
	return &result, err
}

func (b *builder) Download(URI string, state data.Map) (string, error) {
	var resource = url.NewResource(toolbox.URLPathJoin(b.baseURL, URI))
	text, err := resource.DownloadText()
	if err != nil {
		return "", err
	}
	if state != nil {
		text = state.ExpandAsText(text)
	}
	return text, nil

}

func (b *builder) getDeployUploadMap(meta *AppMeta) Map {
	var result = NewMap()
	result.Put("${releasePath}/${app}", "$appPath")
	if len(meta.Assets) == 0 {
		return result
	}
	for _, asset := range meta.Assets {
		result.Put(fmt.Sprintf("${releasePath}/%v", asset), fmt.Sprintf("${appPath}/%v", asset))
	}
	return result
}

func (b *builder) getBuildDownloadMap(meta *AppMeta) Map {
	var result = NewMap()
	if meta.hasAppDirectory {
		result.Put("${buildPath}/app/${app}", "$releasePath")
	} else {
		result.Put("${buildPath}/${app}", "$releasePath")
	}
	if len(meta.Assets) == 0 {
		return result
	}
	for _, asset := range meta.Assets {
		result.Put(fmt.Sprintf("${buildPath}/%v", asset), fmt.Sprintf("${releasePath}%v", asset))
	}
	return result
}

func hasKeyPrefix(keyPrefix string, assets map[string]string) bool {
	for candidate := range assets {
		if strings.HasPrefix(candidate, keyPrefix) {
			return true
		}
	}
	return false
}

func removeComments(assets map[string]string) {
	for k, code := range assets {
		if strings.HasSuffix(k, ".go") && strings.Contains(code, "/*remove") {
			code = strings.Replace(code, "/*remove", "", len(code))
			assets[k] = strings.Replace(code, "remove*/", "", len(code))
		}
	}
}

func (b *builder) buildApp(meta *AppMeta, sdkMeta *SdkMeta, request *RunRequest, assets map[string]string) error {
	buildRequest := request.Build
	var state = data.NewMap()
	var err error
	removeComments(assets)
	request.Build.path = meta.Build
	if meta.UseSdkBuild {
		request.Build.path = sdkMeta.Build
	}

	var buildTemplateURL = toolbox.URLPathJoin(b.baseURL, request.Build.path)
	buildAssets, err := DownloadAll(buildTemplateURL)
	if err != nil {
		return err
	}
	var args = meta.GetArguments(buildRequest.Docker)
	var appFile = fmt.Sprintf("app.yaml")
	var app string
	var appMap Map

	var originURL = meta.OriginURL
	if originURL == "" {
		originURL = request.Origin
	}

	appDirectory := ""
	dependency := ""
	if meta.Dependency != "" {
		dependency = fmt.Sprintf("\n      - %v", strings.Replace(meta.Dependency, "\n", "", strings.Count(meta.Dependency, "\n")))
	}
	if meta.hasAppDirectory {
		appDirectory = "\n      - cd ${buildPath}app"
	}

	state.Put("dependency", dependency)
	state.Put("originURL", fmt.Sprintf(`"%v"`, originURL))
	state.Put("appDirectory", appDirectory)

	if buildRequest.Docker {
		state.Put("args", args)
		appFile = "docker/app.yaml"
		if appMap, err = b.NewAssetMap(buildAssets, appFile, state); err != nil {
			return err
		}

	} else {
		if appMap, err = b.NewAssetMap(buildAssets, "app.yaml", state); err != nil {
			return err
		}
		start := appMap.SubMap("pipeline.start")
		start.Put("arguments", meta.Args)
		appMap.SubMap("pipeline.deploy").Put("upload", b.getDeployUploadMap(meta))
	}

	appMap.SubMap("pipeline.build").Put("download", b.getBuildDownloadMap(meta))

	if app, err = toolbox.AsYamlText(appMap); err != nil {
		return err
	}
	_ = b.UploadToEndly("app.yaml", strings.NewReader(app))

	if buildRequest.Docker {
		var dockerAssets = ""
		if len(meta.Assets) > 0 {
			for _, asset := range meta.Assets {
				if strings.Contains(asset, "config") {
					continue
				}
				if len(dockerAssets) > 0 {
					dockerAssets += "\n"
				}
				parent, _ := path.Split(asset)
				if parent == "" {
					dockerAssets += fmt.Sprintf("ADD %v /", asset)
				} else {
					dockerAssets += fmt.Sprintf("RUN mkdir -p %v\nADD %v /%v", parent, asset, parent)
				}
			}
		}
		state.Put("assets", dockerAssets)
		dockerfile, ok := buildAssets["docker/Dockerfile"]
		if !ok {
			return fmt.Errorf("failed to locate docker file %v", meta.Name)
		}
		dockerfile = state.ExpandAsText(dockerfile)
		_ = b.UploadToEndly("config/Dockerfile", strings.NewReader(dockerfile))
	}
	return err
}

func (b *builder) addSourceCode(meta *AppMeta, request *Build, assets map[string]string) error {

	var dbConfig Map
	if b.registerDb != nil {
		dbConfig = b.registerDb.GetMap("config")
	}
	if meta.DbConfigPath != "" && dbConfig != nil {
		if config, err := b.NewAssetMap(assets, meta.Config, nil); err == nil {
			config.Put(meta.DbConfigPath, dbConfig)
			if YAML, err := toolbox.AsYamlText(config); err == nil {
				assets[meta.Config] = YAML
			}
		}
	}
	for k, v := range assets {
		if k == "meta.yaml" || k == "regression" {
			continue
		}
		b.Upload(k, strings.NewReader(v))
	}
	return nil
}

func (b *builder) Copy(state data.Map, URIs ...string) error {
	for _, URI := range URIs {

		var asset string
		var err error
		if state != nil && path.Ext(URI) == ".json" {
			var JSON = make([]interface{}, 0)
			resource := url.NewResource(toolbox.URLPathJoin(b.baseURL, URI))
			if err = resource.Decode(&JSON); err != nil {
				return err
			}
			expanded := state.Expand(JSON)
			asset, err = toolbox.AsIndentJSONText(expanded)

		} else {
			asset, err = b.Download(URI, state)
		}
		if err != nil {
			return err
		}
		_ = b.UploadToEndly(URI, strings.NewReader(asset))
	}
	return nil
}

func (b *builder) addRun(appMeta *AppMeta, request *RunRequest) error {
	run, err := b.NewMapFromURI("run.yaml", nil)
	if err != nil {
		return err
	}
	var init = run.GetMap("init")
	init.Put("sdk", request.Build.Sdk)
	init.Put("app", request.Build.App)
	if b.dbMeta != nil && b.dbMeta.Credentials != "" {
		var credentialName = b.dbMeta.Credentials
		credentialName = strings.Replace(credentialName, "$", "", 1)
		secret := strings.ToLower(strings.Replace(credentialName, "Credentials", "", 1))
		defaults := run.GetMap("defaults")
		defaults.Put(credentialName, "$"+credentialName)
		run.Put("defaults", defaults)
		init.Put(credentialName, secret)
	}

	if b.dbMeta.Service == "" {

		pieline := run.GetMap("pipeline")
		pielineInit := pieline.GetMap("init")
		pieline.Put("init", pielineInit.Remove("system"))

		pielineDestroy := pieline.GetMap("destroy")
		pieline.Put("destroy", pielineDestroy.Remove("system"))

		run.Put("pipeline", pieline)
	}

	run.Put("init", init)
	if content, err := toolbox.AsYamlText(run); err == nil {
		b.UploadToEndly("run.yaml", strings.NewReader(content))
	}
	return err
}

func (b *builder) NewMapFromURI(URI string, state data.Map) (Map, error) {
	var resource = url.NewResource(toolbox.URLPathJoin(b.baseURL, URI))
	text, err := resource.DownloadText()
	if err != nil {
		return nil, err
	}
	return b.asMap(text, state)
}

func (b *builder) NewAssetMap(assets map[string]string, URI string, state data.Map) (Map, error) {
	value, ok := assets[URI]
	if !ok {
		return nil, fmt.Errorf("unable locate %v, available: %v", URI, toolbox.MapKeysToStringSlice(assets))
	}
	var text = state.ExpandAsText(toolbox.AsString(value))
	return b.asMap(text, state)

}

func (b *builder) buildSystem() error {
	system, err := b.NewMapFromURI("system/system.yaml", nil)
	if err != nil {
		return err
	}
	initMap := system.SubMap("pipeline.init")
	initMap.Put("services", b.services)
	stopImagesMap := system.SubMap("pipeline.destroy.stop-images")
	stopImagesMap.Put("images", b.tags)
	var content string
	if content, err = toolbox.AsYamlText(system); err == nil {
		b.UploadToEndly("system.yaml", strings.NewReader(content))
	}
	return err
}

func (b *builder) buildDatastore() error {
	datastore, err := b.NewMapFromURI("datastore/datastore.yaml", nil)
	if err != nil {
		return err
	}
	pipeline := datastore.SubMap("pipeline")
	pipeline.Put("create-db", b.createDb)
	pipeline.Put("prepare", b.populateDb)
	var content string
	if content, err = toolbox.AsYamlText(datastore); err == nil {
		b.UploadToEndly("datastore.yaml", strings.NewReader(content))
	}
	return err
}

func removeMatchedLines(text string, matchExpr string) string {
	text = strings.Replace(text, "\r", "", len(text))
	var lines = strings.Split(text, "\n")
	var result = make([]string, 0)
	for _, line := range lines {
		if strings.Contains(line, matchExpr) {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func (b *builder) addUseCaseAssets(appMeta *AppMeta, request *RunRequest) error {
	b.Copy(nil,
		"regression/use_cases/001_xx_case/use_case.txt",
		"regression/use_cases/002_yy_case/use_case.txt")
	return nil
}

func (b *builder) buildSeleniumTestAssets(appMeta *AppMeta, request *RunRequest) error {
	b.Copy(nil,
		"regression/req/selenium_init.yaml",
		"regression/req/selenium_destroy.yaml")

	var aMap = map[string]interface{}{
		"in":       "name",
		"output":   "name",
		"expected": "empty",
		"url":      "http://127.0.0.1:8080/",
	}
	if len(appMeta.Selenium) > 0 {
		aMap, _ = util.NormalizeMap(appMeta.Selenium, true)
	}
	test, err := b.Download("regression/selenium_test.yaml", data.Map(aMap))
	if err != nil {
		return err
	}

	b.UploadToEndly("regression/use_cases/001_xx_case/selenium_test.yaml", strings.NewReader(strings.Replace(test, "$index", "1", 2)))
	b.UploadToEndly("regression/use_cases/002_yy_case/selenium_test.yaml", strings.NewReader(strings.Replace(test, "$index", "1", 2)))

	return nil
}

func (b *builder) buildDataTestAssets(appMeta *AppMeta, request *RunRequest) error {

	var setupSource = fmt.Sprintf("regression/%v/setup_data.json", strings.ToLower(b.dbMeta.Kind))
	if request.Datastore.MultiTableMapping {
		setupSource = fmt.Sprintf("regression/%v/v_setup_data.json", strings.ToLower(b.dbMeta.Kind))
	}

	if setupData, err := b.Download(setupSource, nil); err == nil {
		b.UploadToEndly(fmt.Sprintf("regression/use_cases/001_xx_case/%s_data.json", request.Datastore.Name), strings.NewReader(strings.Replace(setupData, "$index", "1", 2)))
		b.UploadToEndly(fmt.Sprintf("regression/use_cases/002_yy_case/%s_data.json", request.Datastore.Name), strings.NewReader(strings.Replace(setupData, "$index", "1", 2)))

		b.UploadToEndly(fmt.Sprintf("regression/%s_data.json", request.Datastore.Name), strings.NewReader("[]"))
		b.UploadToEndly(fmt.Sprintf("regression/%s/data/dummy.json", request.Datastore.Name), strings.NewReader("[]"))

	}
	return nil
}

func (b *builder) buildStaticDataTestAssets(appMeta *AppMeta, request *RunRequest) error {
	var dataSource = "dummy.json"
	if request.Datastore.MultiTableMapping {
		dataSource = "v_dummy.json"
	}
	setupSource := fmt.Sprintf("regression/data/%v", dataSource)
	setupData, err := b.Download(setupSource, nil)
	if err == nil {
		b.UploadToEndly(fmt.Sprintf("regression/%v/data/%v", request.Datastore.Name, dataSource), strings.NewReader(setupData))
	}
	return nil
}

func (b *builder) buildHTTPTestAssets(appMeta *AppMeta, request *RunRequest) error {

	var requestMap = map[string]interface{}{
		"url": "http://127.0.0.1/",
	}
	var expectMap = map[string]interface{}{
		"Code": 200,
	}
	var http map[string]interface{}
	if len(appMeta.HTTP) > 0 {
		http, _ = util.NormalizeMap(appMeta.HTTP, true)
		if value, ok := http["request"]; ok {
			valueMap := toolbox.AsMap(value)
			util.Append(requestMap, valueMap, true)
		}
		if value, ok := http["expect"]; ok {
			valueMap := toolbox.AsMap(value)
			util.Append(expectMap, valueMap, true)
		}
	}

	var httpTest = map[string]interface{}{}
	var httpTestResource = url.NewResource(toolbox.URLPathJoin(b.baseURL, "regression/http_test.json"))
	if err := httpTestResource.Decode(&httpTest); err != nil {
		return err
	}
	var state = data.NewMap()
	state.Put("request", requestMap)
	state.Put("expect", expectMap)
	expandedHttpTest := state.Expand(httpTest)

	if test, err := toolbox.AsIndentJSONText(expandedHttpTest); err == nil {
		b.UploadToEndly("regression/use_cases/001_xx_case/http_test.json", strings.NewReader(strings.Replace(test, "$index", "1", 2)))
		b.UploadToEndly("regression/use_cases/002_yy_case/http_test.json", strings.NewReader(strings.Replace(test, "$index", "1", 2)))
	}

	return nil
}

func (b *builder) buildRESTTestAssets(appMeta *AppMeta, request *RunRequest) error {

	var requestMap = map[string]interface{}{}
	var requesURL = "http://127.0.0.1/"
	var method = "POST"
	var expectMap = map[string]interface{}{}
	var http map[string]interface{}
	if len(appMeta.REST) > 0 {
		http, _ = util.NormalizeMap(appMeta.REST, true)
		if value, ok := http["request"]; ok {
			valueMap := toolbox.AsMap(value)
			util.Append(requestMap, valueMap, true)
		}
		if value, ok := http["expect"]; ok {
			valueMap := toolbox.AsMap(value)
			util.Append(expectMap, valueMap, true)
		}
		if value, ok := http["url"]; ok {
			requesURL = toolbox.AsString(value)
		}
		if value, ok := http["method"]; ok {
			method = toolbox.AsString(value)
		}
	}

	var httpTest = map[string]interface{}{}
	var httpTestResource = url.NewResource(toolbox.URLPathJoin(b.baseURL, "regression/rest_test.json"))
	if err := httpTestResource.Decode(&httpTest); err != nil {
		return err
	}
	var state = data.NewMap()
	state.Put("request", requestMap)
	state.Put("expect", expectMap)
	state.Put("url", requesURL)
	state.Put("method", method)
	state.Put("AsInt", neatly.AsInt)
	state.Put("AsFloat", neatly.AsFloat)
	state.Put("AsBool", neatly.AsBool)

	expandedHttpTest := state.Expand(httpTest)
	if test, err := toolbox.AsIndentJSONText(expandedHttpTest); err == nil {
		b.UploadToEndly("regression/use_cases/001_xx_case/rest_test.json", strings.NewReader(strings.Replace(test, "$index", "1", 2)))
		b.UploadToEndly("regression/use_cases/002_yy_case/rest_test.json", strings.NewReader(strings.Replace(test, "$index", "1", 2)))
	}

	return nil
}

func (b *builder) addRegressionData(appMeta *AppMeta, request *RunRequest) error {
	if request.Datastore == nil {
		return nil
	}
	var state = data.NewMap()
	state.Put("db", request.Datastore.Name)
	dataInit, err := b.NewMapFromURI("datastore/regression/data_init.yaml", state)
	if err != nil {
		return err
	}
	pipeline := dataInit.GetMap("pipeline")
	var prepare Map
	if request.Testing.UseCaseData {
		prepare, err = b.NewMapFromURI("datastore/regression/prepare_data.yaml", state)
	} else {
		prepare, err = b.NewMapFromURI("datastore/regression/prepare.yaml", state)
	}
	if err != nil {
		return err
	}
	var tables interface{} = b.dbMeta.Tables
	if !request.Datastore.MultiTableMapping {
		prepare = prepare.Remove("mapping")
	} else {
		tables = "$tables"
		mappping, err := b.Download("regression/mapping.json", nil)

		if err == nil {
			b.UploadToEndly(fmt.Sprintf("regression/%v/mapping.json", request.Datastore.Name), strings.NewReader(mappping))
		}
	}

	if request.Testing.UseCaseData {
		if !b.dbMeta.Sequence || len(b.dbMeta.Tables) == 0 {
			prepare = prepare.Remove("sequence")
		} else {
			prepare.GetMap("sequence").Put("tables", tables)
		}
		b.buildDataTestAssets(appMeta, request)
	} else {
		b.buildStaticDataTestAssets(appMeta, request)

	}

	pipeline.Put("prepare", prepare)
	pipeline.Put("register", b.registerDb)
	dataYAML, _ := toolbox.AsYamlText(dataInit)
	b.UploadToEndly("regression/req/data.yaml", strings.NewReader(dataYAML))

	return nil
}

func (b *builder) addRegression(appMeta *AppMeta, request *RunRequest) error {
	regression, err := b.Download("regression/regression.csv", nil)
	if err != nil {
		return err
	}
	if err = b.Copy(nil, "regression/var/test_init.json"); err != nil {
		return err
	}

	b.addUseCaseAssets(appMeta, request)

	if request.Testing.Selenium && len(appMeta.Selenium) > 0 {
		b.buildSeleniumTestAssets(appMeta, request)
	} else {
		regression = removeMatchedLines(regression, "selenium")
	}
	if request.Testing.HTTP && len(appMeta.HTTP) > 0 {
		b.buildHTTPTestAssets(appMeta, request)
	} else {
		regression = removeMatchedLines(regression, "HTTP test")
	}

	if request.Testing.REST && len(appMeta.REST) > 0 {
		b.buildRESTTestAssets(appMeta, request)
	} else {
		regression = removeMatchedLines(regression, "REST test")
	}
	var dbMeta = b.dbMeta
	if dbMeta == nil {
		regression = removeMatchedLines(regression, "test data")
	} else {
		b.addRegressionData(appMeta, request)
		if !request.Testing.UseCaseData {
			regression = removeMatchedLines(regression, "setup_data")
		} else {
			regression = strings.Replace(regression, "setup_data", fmt.Sprintf("%v_data", request.Datastore.Name), len(regression))
		}
	}
	b.UploadToEndly("regression/regression.csv", strings.NewReader(regression))
	init, err := b.Download("regression/var/init.json", nil)
	if err != nil {
		return err
	}
	b.UploadToEndly("regression/var/init.json", strings.NewReader(init))

	return nil
}

func (b *builder) URL(URI string) string {
	if b.baseURL == "" {
		return URI
	}
	return toolbox.URLPathJoin(b.baseURL, URI)
}

func (b *builder) UploadToEndly(URI string, reader io.Reader) error {
	URL := toolbox.URLPathJoin(fmt.Sprintf("%vendly", b.destURL), URI)
	content, _ := ioutil.ReadAll(reader)
	//fmt.Printf("%v\n%s\n", URL, content)
	return b.destService.Upload(URL, bytes.NewReader(content))
}

func (b *builder) Upload(URI string, reader io.Reader) error {
	URL := toolbox.URLPathJoin(b.destURL, URI)
	content, _ := ioutil.ReadAll(reader)
	//fmt.Printf("%v\n%s\n", URL, content)
	return b.destService.Upload(URL, bytes.NewReader(content))
}

func newBuilder(baseURL string) *builder {
	return &builder{
		baseURL:     baseURL,
		tags:        make([]string, 0),
		services:    NewMap(),
		createDb:    NewMap(),
		populateDb:  NewMap(),
		destURL:     "mem:///endly/",
		destService: storage.NewPrivateMemoryService(),
		storage:     storage.NewMemoryService(),
	}

}
