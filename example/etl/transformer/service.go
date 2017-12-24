package transformer

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/viant/dsc"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

//Service represents transformer service
type Service interface {
	Copy(request *CopyRequest) *CopyResponse

	TaskList(request *TaskListRequest) *TaskListResponse

	KillTask(request *KillTaskRequest) *KillTaskResponse
}

type service struct {
	mutex *sync.RWMutex
	tasks map[string]*Task
}

func (s *service) registerTask(baseResponse *BaseResponse, taskInfo *TaskInfo, dataset string, request interface{}) {
	var task = &Task{
		ID:           uuid.NewV4().String(),
		Table:        dataset,
		BaseResponse: baseResponse,
		TaskInfo:     taskInfo,
		Request:      request,
	}

	task.Status = "running"
	task.StatusCode = 1
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var now = time.Now()
	for k, v := range s.tasks {
		if v.Expired(now) {
			delete(s.tasks, k)
		}
	}
	s.tasks[task.ID] = task
}

func (s *service) getManager(config *dsc.Config) (dsc.Manager, error) {
	if err := config.Init(); err != nil {
		return nil, err
	}
	return dsc.NewManagerFactory().Create(config)
}

func (s *service) transformIfNeeded(transformer Transformer, source map[string]interface{}) ([]map[string]interface{}, error) {
	if transformer == nil {
		return []map[string]interface{}{source}, nil
	}
	return transformer(source)
}

func (s *service) appendRecords(transformer Transformer, record map[string]interface{}, records []interface{}) error {
	transformed, err := s.transformIfNeeded(transformer, record)
	if err != nil {
		return err
	}
	for _, item := range transformed {
		records = append(records, item)
	}
	return nil
}

func (s *service) persist(destinationManager dsc.Manager, channel chan map[string]interface{}, request *CopyRequest, response *CopyResponse, fetchedCompleted *int32) *sync.WaitGroup {
	var result = &sync.WaitGroup{}
	result.Add(1)
	destination := request.Destination
	tableDescriptor := &dsc.TableDescriptor{
		Table:     destination.Table,
		PkColumns: destination.PkColumns,
		Columns:   destination.Columns,
	}
	var batchSize = request.BatchSize
	if batchSize == 0 {
		batchSize++
	}
	destinationManager.TableDescriptorRegistry().Register(tableDescriptor)

	dmlProvider := dsc.NewMapDmlProvider(tableDescriptor)

	transformer, _ := Transformers[request.Transformer]

	go func() {
		var err error
		defer func() {
			if err != nil {
				s.updateResponse(response, err)
			}
			result.Done()
		}()
		connection, err := destinationManager.ConnectionProvider().Get()
		var fetchCompleted bool
		for {
			var records = make([]interface{}, 0)
			var count = len(channel)
			if count == 0 {
				select {
				case record := <-channel:
					err = s.appendRecords(transformer, record, records)
					if err != nil {
						return
					}
					count = len(channel)
				case <-time.After(time.Millisecond):
					count = len(channel)
					fetchCompleted = atomic.LoadInt32(fetchedCompleted) == 1
					if fetchCompleted {
						if len(records) == 0 && count == 0 {
							return
						}
					} else if len(records) < batchSize {
						continue
					}
				}
			}
			for i := 0; i < count; i++ {
				record := <-channel
				err = s.appendRecords(transformer, record, records)
				if err != nil {
					return
				}
			}
			if len(records) > 0 {
				if request.InsertMode {
					parametrizedSQLProvider := func(item interface{}) *dsc.ParametrizedSQL {
						return dmlProvider.Get(dsc.SQLTypeInsert, item)
					}
					_, err = destinationManager.PersistData(connection, records, request.Destination.Table, dmlProvider, parametrizedSQLProvider)

				} else {
					_, _, err = destinationManager.PersistAll(&records, request.Destination.Table, dmlProvider)
				}
			}
			if err != nil || fetchCompleted {
				return
			}

		}
	}()
	return result
}

func (s *service) copyData(sourceManager, destinationManager dsc.Manager, request *CopyRequest, response *CopyResponse, keys []interface{}) error {
	var batchSize = request.BatchSize
	if batchSize == 0 {
		batchSize = 1
	}
	var records = make(chan map[string]interface{}, batchSize+1)
	var fetchCompleted int32
	waitGroup := s.persist(destinationManager, records, request, response, &fetchCompleted)

	err := sourceManager.ReadAllWithHandler(request.Source.SQL, keys, func(scanner dsc.Scanner) (bool, error) {
		var statusCode = atomic.LoadInt32(&response.StatusCode)
		var record = make(map[string]interface{})
		if statusCode == StatusTaskNotRunning {
			return false, nil
		}
		response.RecordCount++
		err := scanner.Scan(&record)

		if err != nil {
			return false, fmt.Errorf("failed to scan:%v", err)
		}
		if len(record) == 0 {
			response.SkippedRecordCount++
			return true, nil
		}
		records <- record
		return true, nil
	})
	if err != nil {
		return err
	}
	atomic.StoreInt32(&fetchCompleted, 1)
	waitGroup.Wait()
	return nil
}

func (s *service) openKeyFiles(keyPath string) ([]*os.File, error) {
	var result = make([]*os.File, 0)
	directory, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	files, err := directory.Readdir(0)
	if err != nil {
		return nil, err
	}
	for _, info := range files {
		filename := path.Join(keyPath, info.Name())
		osFile, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		result = append(result, osFile)
	}
	return result, nil
}

//Copy copy data from source to destination
func (s *service) Copy(request *CopyRequest) *CopyResponse {
	var response = &CopyResponse{BaseResponse: &BaseResponse{StartTime: time.Now()}, TaskInfo: &TaskInfo{StatusCode: StatusTaskRunning}}
	response.StatusCode = 1
	var dataset = request.Source.Table
	if dataset == "" {
		dataset = request.Source.SQL
	}
	s.registerTask(response.BaseResponse, response.TaskInfo, dataset, request)
	var err error
	var sourceManager, destinationManager dsc.Manager
	defer s.updateResponse(response, err)
	sourceManager, err = s.getManager(request.Source.DsConfig)
	if err != nil {
		return response
	}
	destinationManager, err = s.getManager(request.Destination.DsConfig)
	if err != nil {
		return response
	}
	destinationManager.TableDescriptorRegistry().Register(request.Destination.AsTableDescription())
	keys := []interface{}{}
	err = s.copyData(sourceManager, destinationManager, request, response, keys)
	return response
}

//TaskList returns a list of copy tasks
func (s *service) TaskList(request *TaskListRequest) *TaskListResponse {
	var response = &TaskListResponse{Status: "ok",
		Tasks: make([]*Task, 0),
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, candidate := range s.tasks {
		if candidate.Table == request.Table {
			response.Tasks = append(response.Tasks, candidate)
		}
	}
	return response
}

//KillTask changes status of task to stop it
func (s *service) KillTask(request *KillTaskRequest) *KillTaskResponse {
	var response = &KillTaskResponse{BaseResponse: &BaseResponse{StartTime: time.Now()}}
	for _, candidate := range s.tasks {
		if request.ID == candidate.ID {
			response.Task = candidate
			atomic.StoreInt32(&candidate.StatusCode, StatusTaskNotRunning)
			break
		}
	}
	return response
}

//NewService returns new transformer service
func NewService() Service {
	return &service{
		tasks: make(map[string]*Task),
		mutex: &sync.RWMutex{},
	}
}

func (s *service) updateResponse(response *CopyResponse, err error) {
	atomic.StoreInt32(&response.StatusCode, 0)
	response.EndTime = time.Now()
	if err != nil {
		response.BaseResponse.Status = "error"
		response.Error = err.Error()
	} else if response.BaseResponse.Status == "" {
		response.BaseResponse.Status = "ok"
	}
}