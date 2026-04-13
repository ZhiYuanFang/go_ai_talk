## ADDED Requirements

#### Requirement: Exit marker on DeepSeek exit intent
- When the STT + DeepSeek pipeline indicates the user's intent is to exit the conversation (e.g., explicit text like "我要退出", "结束对话", or a model-level `intent: exit`), the voice backend SHALL return an `exit` boolean set to `true` in the API response.

##### Scenario: Device says "结束对话"
- Given: Device sends audio that transcribes to "结束对话"
- And: DeepSeek returns content or intent indicating exit
- When: Backend processes the request
- Then: Backend response includes `"exit": true`
- And: Backend should not return or synthesize a normal assistant reply audio

##### Notes
- Implementation detail: detection may be based on reply text matching known phrases or a dedicated `intent` field from DeepSeek if available.
- The HTTP and WebSocket responses MUST remain compatible for clients that don't consume the `exit` field.
