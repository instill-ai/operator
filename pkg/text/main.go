package text

import (
	"fmt"
	"sync"

	_ "embed" // embed

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskTokenizerSplitter string = "TASK_TOKENIZER_SPLITTER"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	operator  base.IOperator
)

// Operator is the derived operator
type Operator struct {
	base.Operator
}

// Execution is the derived execution
type Execution struct {
	base.Execution
}

// Init initializes the operator
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

// CreateExecution creates the derived execution
func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)
	return e, nil
}

// Execute executes the derived execution
func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskTokenizerSplitter:

			inputStruct := TokenizerSplitterInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			outputStruct, err := splitTextIntoChunks(inputStruct)
			if err != nil {
				return nil, err
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}
	return outputs, nil
}
