package voice

import (
	"errors"

	"hello/internal/services/aimodel"
)

// mapVoiceLLMError 将 aimodel 闸门/密钥错误映射为 WS/业务层可识别错误。
func mapVoiceLLMError(err error) error {
	if err == nil {
		return nil
	}
	if aimodel.IsQueueFull(err) {
		return &VoiceAIQuotaError{Code: aimodel.CodeLLMQueueFull, Message: err.Error()}
	}
	if errors.Is(err, aimodel.ErrProviderKeyMissing) {
		return StageError{Stage: "chat", Detail: err.Error()}
	}
	return err
}
