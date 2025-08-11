package util

import (
	"errors"
	"github.com/alibabacloud-go/tea/tea"
	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetErrorMessage(t *testing.T) {
	errMsg := "test message"
	sdkErr := &tea.SDKError{
		Message: tea.String(errMsg),
	}
	serverErr := sdkerrors.NewServerError(http.StatusOK, errMsg, "")

	assert.Equal(t, errMsg, GetErrorMessage(sdkErr))
	assert.Equal(t, errMsg, GetErrorMessage(serverErr))
	assert.Equal(t, errMsg, GetErrorMessage(errors.New(errMsg)))
	assert.Equal(t, "", GetErrorMessage(nil))
}

func TestIsThrottlingError(t *testing.T) {
	sdkThrottlingError := &tea.SDKError{
		Code:    tea.String("Throttling.User"),
		Message: tea.String("Request was denied due to user flow control."),
	}
	serverThrottlingError := sdkerrors.NewServerError(http.StatusOK,
		"{\"Code\": \"Throttling.User\", \"Message\": \"Request was denied due to user flow control.\"}", "")

	sdkError := &tea.SDKError{
		Code:    tea.String("Test.Error"),
		Message: tea.String("test error"),
	}

	serverError := sdkerrors.NewServerError(http.StatusOK,
		"{\"Code\": \"Test.Error\", \"Message\": \"test error\"}", "")

	assert.Equal(t, true, IsThrottlingError(sdkThrottlingError))
	assert.Equal(t, true, IsThrottlingError(serverThrottlingError))
	assert.Equal(t, false, IsThrottlingError(sdkError))
	assert.Equal(t, false, IsThrottlingError(serverError))
	assert.Equal(t, false, IsThrottlingError(nil))
}
