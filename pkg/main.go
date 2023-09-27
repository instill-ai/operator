package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
	"github.com/instill-ai/operator/pkg/json"
	"github.com/instill-ai/operator/pkg/pipeline"
	"github.com/instill-ai/operator/pkg/rest"
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
	restOperator           base.IOperator
}

type OperatorOptions struct {
	Base64         base64.OperatorOptions
	TextExtraction textextraction.OperatorOptions
	Pipeline       pipeline.OperatorOptions
	JSON           json.OperatorOptions
	REST           rest.OperatorOptions
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		base64Operator := base64.Init(logger, options.Base64)
		textExtractionOperator := textextraction.Init(logger, options.TextExtraction)
		pipelineOperator := pipeline.Init(logger, options.Pipeline)
		jsonOperator := json.Init(logger, options.JSON)
		restOperator := rest.Init(logger, options.REST)

		operator = &Operator{
			BaseOperator:           base.BaseOperator{Logger: logger},
			base64Operator:         base64Operator,
			textExtractionOperator: textExtractionOperator,
			pipelineOperator:       pipelineOperator,
			jsonOperator:           jsonOperator,
			restOperator:           restOperator,
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
		for _, uid := range restOperator.ListOperatorDefinitionUids() {
			def, err := restOperator.GetOperatorDefinitionByUid(uid)
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
	case o.restOperator.HasUid(defUid):
		return o.restOperator.CreateExecution(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no operator uid: %s", defUid)
	}
}
