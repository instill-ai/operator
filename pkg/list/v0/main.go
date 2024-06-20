package list

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskFilter = "TASK_FILTER"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	op   base.IOperator
)

type operator struct {
	base.Operator
}

type execution struct {
	base.Execution
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

// Init returns an implementation of IOperator that processes JSON objects.
func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		op = &operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := op.LoadOperatorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return op
}

func (o *operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &execution{}

	switch task {
	case taskFilter:
		e.execute = e.filter
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)

	return e, nil
}

func (e *execution) filter(_ /* in */ *structpb.Struct) (*structpb.Struct, error) {
	return nil, errmsg.AddMessage(fmt.Errorf("unsupported task"), "The filter task is not available yet for execution.")
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
