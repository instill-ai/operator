package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/operator/pkg/json"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
	"github.com/instill-ai/operator/pkg/pipeline"
	"github.com/instill-ai/operator/pkg/textextraction"
)

var (
	once     sync.Once
	operator base.IOperator
)

type Operator struct {
	base.BaseOperator
	base64Operator         base.IOperator
	textExtractionOperator base.IOperator
	pipelineOperator       base.IOperator
	jsonOperator           base.IOperator
}

type OperatorOptions struct {
	Base64         base64.OperatorOptions
	TextExtraction textextraction.OperatorOptions
	Pipeline       pipeline.OperatorOptions
	JSON           json.OperatorOptions
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		base64Operator := base64.Init(logger, options.Base64)
		textExtractionOperator := textextraction.Init(logger, options.TextExtraction)
		pipelineOperator := pipeline.Init(logger, options.Pipeline)
		jsonOperator := json.Init(logger, options.JSON)

		operator = &Operator{
			BaseOperator:           base.BaseOperator{Logger: logger},
			base64Operator:         base64Operator,
			textExtractionOperator: textExtractionOperator,
			pipelineOperator:       pipelineOperator,
			jsonOperator:           jsonOperator,
		}
		for _, uid := range base64Operator.ListOperatorDefinitionUids() {
			def, err := base64Operator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
		for _, uid := range textExtractionOperator.ListOperatorDefinitionUids() {
			def, err := textExtractionOperator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
		for _, uid := range pipelineOperator.ListOperatorDefinitionUids() {
			def, err := pipelineOperator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
		for _, uid := range jsonOperator.ListOperatorDefinitionUids() {
			def, err := jsonOperator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	})
	return operator
}

func (o *Operator) CreateExecution(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	switch {
	case o.base64Operator.HasUid(defUid):
		return o.base64Operator.CreateExecution(defUid, config, logger)
	case o.textExtractionOperator.HasUid(defUid):
		return o.textExtractionOperator.CreateExecution(defUid, config, logger)
	case o.pipelineOperator.HasUid(defUid):
		return o.pipelineOperator.CreateExecution(defUid, config, logger)
	case o.jsonOperator.HasUid(defUid):
		return o.jsonOperator.CreateExecution(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no operator uid: %s", defUid)
	}
}
