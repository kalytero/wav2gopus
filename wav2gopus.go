package wav2gopus

import (
	"fmt"

	opusv2 "gopkg.in/hraban/opus.v2"
)

const wav_header_size = 44

// const sampleRate = 24000 // VOICEVOX SampleRate
// const channels = 1       // VOICEVOX Audio Channel
// const frameSizeMs = 20
// const frameSize = sampleRate * frameSizeMs / 1000

func Encoder(adapter *opusv2.Encoder, channel *chan []byte, wav []byte, opusFrameSize int) error {
	var length int
	var pcm []int16
	var wav_bytes []byte

	var opus_bytes []byte
	var opus_frame []int16
	var opus_frame_len int
	var end int

	var err error

	// wav to pcm
	length = len(wav) - wav_header_size
	if length < 0 {
		return fmt.Errorf("bytes are not wav")
	}
	if length%2 == 1 {
		return fmt.Errorf("wav must be 2 channel")
	}
	length = length / 2
	pcm = make([]int16, length)
	wav_bytes = wav[44:]
	for i := 0; i < length; i++ {
		pcm[i] = int16(wav_bytes[i*2]) + int16(wav_bytes[i*2+1])<<8
	}

	// pcm to opus & send to channel
	opus_bytes = make([]byte, opusFrameSize*2)
	for i := 0; i < length; i += opusFrameSize {
		end = i + opusFrameSize
		if i+opusFrameSize > length {
			break
		}
		opus_frame = pcm[i:end]
		opus_frame_len, err = adapter.Encode(opus_frame, opus_bytes)
		if err != nil {
			return err
		}
		*channel <- opus_bytes[:opus_frame_len]
	}
	return nil
}
