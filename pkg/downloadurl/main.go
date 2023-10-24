package downloadurl

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/pkg/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

const (
	download = "TASK_DOWNLOAD_BASE64"
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

type Base64Download struct {
	URL  string `json:"url"`
	Data string `json:"data"`
}

func DownloadAsBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting image: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		output := structpb.Struct{}
		switch e.Task {
		case download:
			downloadInput := Base64Download{}
			err := base.ConvertFromStructpb(input, &downloadInput)
			if err != nil {
				return nil, err
			}

			downloadInput.Data, err = DownloadAsBase64(downloadInput.URL)
			if err != nil {
				return nil, err
			}
			outputJson, err := json.Marshal(downloadInput)
			if err != nil {
				return nil, err
			}
			err = protojson.Unmarshal(outputJson, &output)
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
		outputs = append(outputs, &output)
	}
	return outputs, nil
}
