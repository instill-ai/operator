package json

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

	om "github.com/instill-ai/component/pkg/objectmapper"
)

const (
	getValue = "TASK_GET_VALUE"
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

type GetValueInput struct {
	Path       string `json:"path"`
	JSONString string `json:"json_string"`
}

type GetValueRes struct {
	Result any `json:"result"`
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

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		output := structpb.Struct{}
		switch e.Task {
		case getValue:
			getValueStruct := GetValueInput{}
			err := base.ConvertFromStructpb(input, &getValueStruct)
			if err != nil {
				return nil, err
			}
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(getValueStruct.JSONString), &obj)
			if err != nil {
				return nil, err
			}
			res, err := om.GetSrcValueByTag(obj, getValueStruct.Path)
			if err != nil {
				return nil, err
			}
			outputJson, err := json.Marshal(GetValueRes{Result: res})
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
