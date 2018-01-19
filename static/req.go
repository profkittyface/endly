package static

import (
	"bytes"
	"github.com/viant/toolbox/storage"
	"log"
)

func init() {
	var memStorage = storage.NewMemoryService();
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/notify_error.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,110,111,116,105,102,121,95,101,114,111,114,34,44,10,32,32,34,84,97,115,107,115,34,58,34,110,111,116,105,102,121,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,32,32,34,83,77,84,80,82,101,115,111,117,114,99,101,34,58,34,36,83,77,84,80,82,101,115,111,117,114,99,101,34,44,10,32,32,32,32,32,32,36,97,114,103,115,48,10,32,32,125,10,125,10,10,10,10}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/notify_error.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/mysql.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,100,111,99,107,101,114,105,122,101,100,95,109,121,115,113,108,34,44,10,32,32,34,84,97,115,107,115,34,58,34,36,116,97,115,107,115,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,34,115,116,111,112,83,121,115,116,101,109,77,121,115,113,108,34,58,32,116,114,117,101,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,34,58,34,36,116,97,114,103,101,116,72,111,115,116,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,34,58,32,34,99,111,110,102,105,103,47,109,121,46,99,110,102,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,109,121,115,113,108,67,114,101,100,101,110,116,105,97,108,34,58,34,36,109,121,115,113,108,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,115,101,114,118,105,99,101,73,110,115,116,97,110,99,101,78,97,109,101,34,58,32,34,36,105,110,115,116,97,110,99,101,34,44,10,32,32,32,32,34,109,121,115,113,108,86,101,114,115,105,111,110,34,58,34,36,109,121,115,113,108,86,101,114,115,105,111,110,34,44,10,32,32,32,32,34,101,120,112,111,114,116,70,105,108,101,34,58,34,36,101,120,112,111,114,116,70,105,108,101,34,44,10,32,32,32,32,34,105,109,112,111,114,116,70,105,108,101,34,58,34,36,105,109,112,111,114,116,70,105,108,101,34,10,32,32,125,10,125}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/mysql.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/ec2.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,101,99,50,34,44,10,32,32,34,84,97,115,107,115,34,58,32,34,36,116,97,115,107,115,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,34,97,119,115,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,97,119,115,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,101,99,50,73,110,115,116,97,110,99,101,73,100,34,58,32,34,101,99,50,73,110,115,116,97,110,99,101,73,100,34,10,32,32,125,10,125,10,10}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/ec2.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/set_go.json", bytes.NewReader([]byte{123,10,32,32,34,84,97,114,103,101,116,34,58,32,123,10,32,32,32,32,34,85,82,76,34,58,32,34,115,115,104,58,47,47,36,123,116,97,114,103,101,116,72,111,115,116,125,47,34,44,10,32,32,32,32,34,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,10,32,32,125,44,10,32,32,34,83,100,107,34,58,32,34,103,111,34,44,10,32,32,34,86,101,114,115,105,111,110,34,58,32,34,36,103,111,86,101,114,115,105,111,110,34,44,10,32,32,34,69,110,118,34,58,123,10,32,32,32,32,34,71,79,80,65,84,72,34,58,34,36,71,79,80,65,84,72,34,10,32,32,125,10,125}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/set_go.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/README.txt", bytes.NewReader([]byte{84,104,105,115,32,100,105,114,101,99,116,111,114,121,32,115,116,111,114,101,115,32,119,111,114,107,102,108,111,119,32,114,117,110,32,114,101,113,117,101,115,116,32,116,104,97,116,32,99,97,110,32,98,101,32,117,115,101,100,32,97,115,32,114,101,102,101,114,101,110,99,101,32,105,110,32,97,110,121,32,119,111,114,107,102,108,111,119,32,105,102,32,116,104,101,32,108,111,99,97,108,32,114,101,113,32,119,105,116,104,32,116,104,101,32,115,97,109,101,32,110,97,109,101,32,100,111,101,115,32,110,111,116,32,101,120,105,115,116,46}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/README.txt %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/set_jdk.json", bytes.NewReader([]byte{123,10,32,32,34,84,97,114,103,101,116,34,58,32,123,10,32,32,32,32,34,85,82,76,34,58,32,34,115,115,104,58,47,47,36,123,116,97,114,103,101,116,72,111,115,116,125,47,34,44,10,32,32,32,32,34,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,10,32,32,125,44,10,32,32,34,83,100,107,34,58,32,34,106,100,107,34,44,10,32,32,34,86,101,114,115,105,111,110,34,58,32,34,36,106,100,107,86,101,114,115,105,111,110,34,10,125}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/set_jdk.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/memcached.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,100,111,99,107,101,114,105,122,101,100,95,109,101,109,99,97,99,104,101,100,34,44,10,32,32,34,84,97,115,107,115,34,58,34,36,116,97,115,107,115,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,34,117,114,108,34,58,32,34,115,99,112,58,47,47,36,123,116,97,114,103,101,116,72,111,115,116,125,47,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,34,58,34,36,116,97,114,103,101,116,72,111,115,116,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,34,58,32,34,99,111,110,102,105,103,47,97,101,114,111,115,112,105,107,101,46,99,111,110,102,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,115,101,114,118,105,99,101,73,110,115,116,97,110,99,101,78,97,109,101,34,58,32,34,36,105,110,115,116,97,110,99,101,34,44,10,32,32,32,32,34,109,97,120,77,101,109,111,114,121,34,58,34,36,109,97,120,77,101,109,111,114,121,34,10,32,32,125,10,125,10}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/memcached.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/aerospike.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,100,111,99,107,101,114,105,122,101,100,95,97,101,114,111,115,112,105,107,101,34,44,10,32,32,34,84,97,115,107,115,34,58,34,36,116,97,115,107,115,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,34,58,34,36,116,97,114,103,101,116,72,111,115,116,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,34,58,32,34,99,111,110,102,105,103,47,97,101,114,111,115,112,105,107,101,46,99,111,110,102,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,115,101,114,118,105,99,101,73,110,115,116,97,110,99,101,78,97,109,101,34,58,32,34,36,105,110,115,116,97,110,99,101,34,10,32,32,125,10,125,10}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/aerospike.json %v", err)
		}
	}
	{
		err := memStorage.Upload("mem://github.com/viant/endly/req/tomcat.json", bytes.NewReader([]byte{123,10,32,32,34,78,97,109,101,34,58,32,34,116,111,109,99,97,116,34,44,10,32,32,34,84,97,115,107,115,34,58,34,36,116,97,115,107,115,34,44,10,32,32,34,80,97,114,97,109,115,34,58,32,123,10,32,32,32,32,34,97,112,112,34,58,32,34,36,97,112,112,34,44,10,32,32,32,32,34,99,97,116,97,108,105,110,97,79,112,116,115,34,58,32,34,36,99,97,116,97,108,105,110,97,79,112,116,115,34,44,10,32,32,32,32,34,106,100,107,86,101,114,115,105,111,110,34,58,32,34,36,106,100,107,86,101,114,115,105,111,110,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,34,44,10,32,32,32,32,34,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,58,32,34,36,116,97,114,103,101,116,72,111,115,116,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,97,112,112,68,105,114,101,99,116,111,114,121,34,58,32,34,36,123,97,112,112,82,111,111,116,68,105,114,101,99,116,111,114,121,125,47,36,123,97,112,112,125,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,114,108,34,58,32,34,99,111,110,102,105,103,47,116,111,109,99,97,116,45,115,101,114,118,101,114,46,120,109,108,34,44,10,32,32,32,32,34,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,58,34,36,99,111,110,102,105,103,85,82,76,67,114,101,100,101,110,116,105,97,108,34,44,10,32,32,32,32,34,116,111,109,99,97,116,80,111,114,116,34,58,32,34,36,112,111,114,116,34,44,10,32,32,32,32,34,116,111,109,99,97,116,83,104,117,116,100,111,119,110,80,111,114,116,34,58,32,34,36,107,105,108,108,80,111,114,116,34,44,10,32,32,32,32,34,116,111,109,99,97,116,65,74,80,67,111,110,110,101,99,116,111,114,80,111,114,116,34,58,32,34,36,99,111,110,110,101,99,116,111,114,80,111,114,116,34,44,10,32,32,32,32,34,116,111,109,99,97,116,82,101,100,105,114,101,99,116,80,111,114,116,34,58,32,34,36,114,101,100,105,114,101,99,116,80,111,114,116,34,44,10,32,32,32,32,34,116,111,109,99,97,116,86,101,114,115,105,111,110,34,58,34,36,116,111,109,99,97,116,86,101,114,115,105,111,110,34,44,10,32,32,32,32,34,102,111,114,99,101,68,101,112,108,111,121,34,58,34,36,102,111,114,99,101,68,101,112,108,111,121,34,44,10,32,32,32,32,34,106,112,100,97,65,100,100,114,101,115,115,34,58,34,36,106,112,100,97,65,100,100,114,101,115,115,34,10,32,32,125,10,125,10,10}))
		if err != nil {
			log.Printf("failed to upload: mem://github.com/viant/endly/req/tomcat.json %v", err)
		}
	}
}
