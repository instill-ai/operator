package rest

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
)

const (
	callEndpoint = "TASK_CALL_ENDPOINT"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	operator  base.IOperator
)

type OperatorOptions struct{}

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
		var output *structpb.Struct
		switch e.Task {
		case callEndpoint:
			req := Request{}
			err := base.ConvertFromStructpb(input, &req)
			if err != nil {
				return nil, err
			}
			resp, err := req.sendReq()
			if err != nil {
				return nil, err
			}
			output, err = responseToStruct(resp)
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

func responseToStruct(res Response) (*structpb.Struct, error) {
	s := &structpb.Struct{}
	outputJSON, err := json.Marshal(res)
	if err != nil {
		return s, err
	}
	err = protojson.Unmarshal(outputJSON, s)
	if err != nil {
		return s, err
	}
	//handle JSON body
	if json.Valid([]byte(res.ResponseBody)) {
		var jsonBody any
		err = json.Unmarshal([]byte(res.ResponseBody), &jsonBody)
		if err == nil {
			str := structpb.Struct{}
			err = protojson.Unmarshal([]byte(res.ResponseBody), &str)
			s.Fields["response_body"] = structpb.NewStructValue(&str)
		}
	}
	return s, err
}
