package surveyresponses

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	studydefinition "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
	studytypes "github.com/case-framework/case-backend/pkg/types/study"
)

func valueToStr(resultVal interface{}) string {
	if resultVal == nil {
		return ""
	}

	var str string
	switch colValue := resultVal.(type) {
	case string:
		str = colValue
	case int:
		str = fmt.Sprintf("%d", colValue)
	case int64:
		str = fmt.Sprintf("%d", colValue)
	case float64:
		str = fmt.Sprintf("%f", colValue)
	case *studytypes.ResponseItem:
		jsonBytes, err := json.Marshal(colValue)
		if err != nil {
			slog.Debug("error while parsing response column", slog.String("error", err.Error()))
			return err.Error()
		}
		str = string(jsonBytes)
	}
	return str
}

func findResponse(responses []studytypes.SurveyItemResponse, key string) *studytypes.SurveyItemResponse {
	for _, r := range responses {
		if r.Key == key {
			return &r
		}
	}
	return nil
}

func getResponseColNamesForAllVersions(
	surveyVersions []studydefinition.SurveyVersionPreview,
	questionOptionSep string,
) []string {
	colNames := map[string]bool{}
	for _, version := range surveyVersions {
		for _, question := range version.Questions {
			newColNames := getResponseColNamesForQuestion(question, questionOptionSep)
			for _, colName := range newColNames {
				colNames[colName] = true
			}
		}
	}

	uniqueColNames := []string{}
	for colName := range colNames {
		uniqueColNames = append(uniqueColNames, colName)
	}

	return uniqueColNames
}

func getResponseColumns(
	question studydefinition.SurveyQuestion,
	response *studytypes.SurveyItemResponse,
	questionOptionSep string,
) map[string]interface{} {
	qTypeHandl, ok := questionTypeHandlers[question.QuestionType]
	if !ok {
		slog.Error("no handler found for question type", slog.String("questionType", question.QuestionType))
		return map[string]interface{}{}
	}
	return qTypeHandl.ParseResponse(question, response, questionOptionSep)
}

func getResponseColNamesForQuestion(
	question studydefinition.SurveyQuestion,
	questionOptionSep string,
) []string {
	qTypeHandl, ok := questionTypeHandlers[question.QuestionType]
	if !ok {
		slog.Error("no handler found for question type", slog.String("questionType", question.QuestionType))
		return []string{}
	}
	return qTypeHandl.GetResponseColumnNames(question, questionOptionSep)
	/*

	   case studydefinition.QUESTION_TYPE_DROPDOWN:

	   	return processResponseForSingleChoice(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_LIKERT:

	   	return processResponseForSingleChoice(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_LIKERT_GROUP:

	   	return handleSingleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY:

	   	return processResponseForSingleChoice(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY:

	   	return processResponseForSingleChoice(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_TEXT_INPUT:

	   	return processResponseForInputs(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_DATE_INPUT:

	   	return processResponseForInputs(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_NUMBER_INPUT:

	   	return processResponseForInputs(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_NUMERIC_SLIDER:

	   	return processResponseForInputs(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_EQ5D_SLIDER:

	   	return processResponseForInputs(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_RESPONSIVE_TABLE:

	   	return processResponseForResponsiveTable(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_MATRIX:

	   	return processResponseForMatrix(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_CLOZE:

	   	return processResponseForCloze(question, response, questionOptionSep)

	   case studydefinition.QUESTION_TYPE_UNKNOWN:

	   	return processResponseForUnknown(question, response, questionOptionSep)

	   default:

	   		return []string{}
	   	}
	*/
}

func retrieveResponseItem(response *studytypes.SurveyItemResponse, fullKey string) *studytypes.ResponseItem {
	if response == nil || response.Response == nil {
		return nil
	}
	keyParts := strings.Split(fullKey, ".")

	var result *studytypes.ResponseItem
	for _, key := range keyParts {
		if result == nil {
			if key != response.Response.Key {
				return nil
			}
			result = response.Response
			continue
		}
		found := false
		for _, item := range result.Items {
			if item.Key == key {
				result = item
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return result
}

func retrieveResponseItemByShortKey(response *studytypes.SurveyItemResponse, shortKey string) *studytypes.ResponseItem {
	if response == nil || response.Response == nil {
		return nil
	}

	var result *studytypes.ResponseItem
	if response.Response.Key == shortKey {
		return response.Response
	}

	result = response.Response

	for _, item := range result.Items {
		if item.Key == shortKey {
			return item
		}
	}

	for _, item := range result.Items {
		res := retrieveResponseItemByShortKey(&studytypes.SurveyItemResponse{
			Response: item,
		}, shortKey)
		if res != nil {
			return res
		}
	}
	return nil
}
