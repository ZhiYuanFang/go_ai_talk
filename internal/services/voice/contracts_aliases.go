package voice

import contracts "hello/internal/services/contracts"

type AudioMeta = contracts.AudioMeta
type StageError = contracts.StageError
type StreamASRSession = contracts.StreamASRSession
type STTProfile = contracts.STTProfile
type StreamTTSSession = contracts.StreamTTSSession

const (
	STTProfileChat      = contracts.STTProfileChat
	STTProfileDictation = contracts.STTProfileDictation
)
type Contract = contracts.VoiceContract
type VoiceContract = contracts.VoiceContract
