package wav2gopus

import (
	"fmt"

	Macro "github.com/seyukun/gomacros"
	Opus "gopkg.in/hraban/opus.v2"
)

func Encoder(wav_raw []byte, opus_dst *chan []byte, opus_sample_rate int) error {
	var wav_len int
	var wav_sample_rate int
	var pcm_len int
	var pcm_i16 []int16
	var err error

	// wav validation
	wav_len = len(wav_raw) - 44
	if wav_len < 0 {
		return fmt.Errorf("bytes are not wav")
	}

	// wav analysis
	wav_sample_rate = int(uint16(wav_raw[24]) | uint16(wav_raw[25])<<8)

	// wav to pcm
	pcm_i16, pcm_len = wav_to_pcm(wav_raw)
	pcm_i16, pcm_len = pcm_resample(pcm_i16, pcm_len, wav_sample_rate, opus_sample_rate)
	pcm_i16, pcm_len = pcm_mono_to_streo(pcm_i16, pcm_len)

	// pcm to opus & send to channel
	err = pcm_to_opus(pcm_i16, pcm_len, opus_dst, opus_sample_rate)
	if err != nil {
		return fmt.Errorf("%s(%d) %s", Macro.M__FILE__(), Macro.M__LINE__(), err.Error())
	}
	return nil
}

func wav_to_pcm(wav_raw []byte) ([]int16, int) {
	var wav_len int
	var wav_ui8 []uint8
	var pcm_len int
	var pcm_i16 []int16

	wav_len = len(wav_raw) - 44
	wav_ui8 = wav_raw[44:]
	pcm_len = wav_len / 2
	pcm_i16 = make([]int16, pcm_len)
	for i := 0; i < pcm_len; i++ {
		pcm_i16[i] = int16(wav_ui8[i*2]) + (int16(wav_ui8[i*2+1]) << 8)
	}
	return pcm_i16, pcm_len
}

func pcm_resample(src_pcm []int16, pcm_len, src_rate, dst_rate int) ([]int16, int) {
	var rate_ratio float64
	var dst_pcm []int16
	var src_index float64
	var l_index int
	var r_index int
	var l_value float64
	var r_value float64
	var interpolation float64

	if (src_rate == dst_rate) || (pcm_len == 0) {
		return src_pcm, pcm_len
	}
	rate_ratio = float64(dst_rate) / float64(src_rate)
	dst_pcm = make([]int16, int(float64(pcm_len)*rate_ratio))
	for i := 0; i < len(dst_pcm); i++ {
		src_index = float64(i) / rate_ratio
		l_index = int(src_index)
		r_index = l_index + 1
		if r_index >= pcm_len {
			dst_pcm[i] = src_pcm[l_index]
		} else {
			l_value = float64(src_pcm[l_index])
			r_value = float64(src_pcm[r_index])
			interpolation = src_index - float64(l_index)
			dst_pcm[i] = int16(l_value*(1-interpolation) + r_value*interpolation)
		}
	}
	return dst_pcm, len(dst_pcm)
}

func pcm_to_opus(pcm_i16 []int16, pcm_len int, opus_dst *chan []byte, opus_sample_rate int) error {
	var opus_encoder *Opus.Encoder
	var opus_frame_size int
	var opus_ui8 []uint8
	var opus_frame []int16
	var opus_frame_len int
	var end int
	var err error

	opus_encoder, err = Opus.NewEncoder(opus_sample_rate, 2, Opus.AppRestrictedLowdelay)
	opus_encoder.SetBitrateToMax()
	if err != nil {
		return fmt.Errorf("%s(%d) %s", Macro.M__FILE__(), Macro.M__LINE__(), err.Error())
	}
	opus_frame_size = opus_sample_rate * 20 / 1000 * 2
	opus_ui8 = make([]byte, opus_frame_size*2)
	for i := 0; i < pcm_len; i += opus_frame_size {
		if i+opus_frame_size > pcm_len {
			break
		}
		end = i + opus_frame_size
		opus_frame = pcm_i16[i:end]
		opus_frame_len, err = opus_encoder.Encode(opus_frame, opus_ui8)
		if err != nil {
			return fmt.Errorf("%d: %s", Macro.M__LINE__(), err.Error())
		}
		*opus_dst <- opus_ui8[:opus_frame_len]
	}
	return nil
}

func pcm_mono_to_streo(pcm_mono []int16, pcm_len int) ([]int16, int) {
	var pcm_stereo_len = pcm_len * 4
	var pcm_stereo = make([]int16, pcm_stereo_len)

	for i := 0; i < pcm_len; i++ {
		sample := pcm_mono[i]
		pcm_stereo[i*2] = sample
		pcm_stereo[i*2+1] = sample
	}

	return pcm_stereo, len(pcm_stereo)
}
