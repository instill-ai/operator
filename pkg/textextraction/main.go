package textextraction

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
)

const (
	taskExtractFromPath = "TASK_EXTRACT_FROM_PATH"
	taskExtractFromFile = "TASK_EXTRACT_FROM_FILE"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	operator  base.IOperator
)

type Operator struct {
	base.Operator
}

type Execution struct {
	base.Execution
}

type FromPath struct {
	FilePath string `json:"path"`
}

type FromFile struct {
	FileContents string `json:"file_contents"`
	ContentType  string `json:"content_type"`
}

type Output struct {
	Text string `json:"text"`
}

func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := operator.LoadOperatorDefinitions(definitionsJSON, tasksJSON, nil)
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

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		op := Output{}
		switch e.Task {
		case taskExtractFromPath:
			obj := FromPath{}
			err := base.ConvertFromStructpb(input, &obj)
			if err != nil {
				return nil, err
			}
			op.Text, err = PathToText(obj.FilePath)
			if err != nil {
				return nil, err
			}
		case taskExtractFromFile:
			obj := FromFile{}
			err := base.ConvertFromStructpb(input, &obj)
			if err != nil {
				return nil, err
			}
			obj.FileContents, err = base64.Decode(obj.FileContents)
			if err != nil {
				return nil, err
			}
			op.Text, err = BytesToText([]byte(obj.FileContents), obj.ContentType)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
		outputJson, err := json.Marshal(op)
		if err != nil {
			return nil, err
		}
		output := structpb.Struct{}
		err = protojson.Unmarshal(outputJson, &output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, &output)
	}

	return outputs, nil
}
