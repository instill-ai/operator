package json

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	marshal   = "TASK_MARSHAL"
	unmarshal = "TASK_UNMARSHAL"
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
		output := structpb.Struct{Fields: make(map[string]*structpb.Value)}
		switch e.Task {
		case marshal:
			b, err := protojson.Marshal(input.Fields["object"])
			if err != nil {
				return nil, err
			}
			output.Fields["string"] = structpb.NewStringValue(string(b))
		case unmarshal:
			obj := structpb.Struct{}
			err := protojson.Unmarshal([]byte(input.Fields["string"].GetStringValue()), &obj)
			if err != nil {
				return nil, err
			}
			output.Fields["object"] = structpb.NewStructValue(&obj)

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
		outputs = append(outputs, &output)
	}

	return outputs, nil
}
