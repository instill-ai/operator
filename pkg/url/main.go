package download

import (
	_ "embed"
	"fmt"	
	"github.com/instill-ai/component/pkg/base"
	"google.golang.org/protobuf/types/known/structpb"
	"sync"

	"io/ioutil"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"encoding/base64"
	"net/http"
)

const (
	download = "DOWNLOAD"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	operator  base.IOperator
	once      sync.Once
)

type Operator struct {
	base.Operator
}

type Execution struct {
	base.Execution
}

func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := operator.LoadOperatorDefinitions(definitionsJSON, tasksJSON)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return operator
}

func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)
	return e, nil
}


type DownloadInput struct {
	URL         string            `json:"url"`
}

func DownloadImageAsBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting image: %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case download:
			downloadInput := DownloadInput{}
			err := base.ConvertFromStructpb(input, &downloadInput)
			if err != nil {
				return nil, err
			}
			
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}