package siop

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	Protocol = 5

	KeyPlaceholder = "_placeholder"
	KeyNum         = "num"
)

var (
	ErrEmptyData          = errors.New("empty data")
	ErrInvalidType        = errors.New("invalid packet type")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrIllegalAttachments = errors.New("illegal attachments")
	ErrReconstructPacket  = errors.New("got binary data when not reconstructing a packet")
)

var _ Encoder = (*encoder)(nil)
var _ Decoder = (*decoder)(nil)

type (
	Encoder interface {
		Encode(pkt Packet) []any
	}

	Decoder interface {
		Add(data any) error
		Destroy()
		OnDecoded(callback Callback)
	}
)

type (
	encoder struct{}

	reconstructor struct {
		pkt     *Packet
		buffers []any
	}

	Callback func(pkt Packet) error

	decoder struct {
		reconstructor *reconstructor
		callback      Callback
	}

	deconstructor struct {
		buffers []any
	}
)

func ParseTypeASCII(r uint8) (PacketType, error) {
	return ParseType(string(r))
}

func ParseType(str string) (PacketType, error) {
	if len(str) > 1 {
		if p, ok := StringToPacketType[str]; ok {
			return p, nil
		}
		return Connect, fmt.Errorf("%s error: %w", str, ErrInvalidType)
	}

	n, err := strconv.Atoi(str)
	if err != nil || n < Connect.Int() || n > BinaryAck.Int() {
		return Connect, fmt.Errorf("%s error: %w", str, ErrInvalidType)
	}

	return PacketType(n), nil
}

func Encode(pkt Packet) []any {
	return NewEncoder().Encode(pkt)
}

func NewEncoder() *encoder {
	return new(encoder)
}

func (e *encoder) Encode(pkt Packet) []any {
	switch pkt.Type {
	case Event, Ack:
		if e.hasBinary(pkt.Data) {
			if pkt.Type == Event {
				pkt.Type = BinaryEvent
			} else {
				pkt.Type = BinaryAck
			}
			return e.encodeAsBinary(pkt)
		}
	}
	return []any{e.encodeAsString(pkt)}
}

func (e *encoder) encodeAsBinary(pkt Packet) []any {
	buffers := new(deconstructor).do(&pkt)
	return append([]any{e.encodeAsString(pkt)}, buffers...)
}

func (e *encoder) encodeAsString(pkt Packet) string {
	str := pkt.Type.Encode()

	if pkt.Type == BinaryEvent || pkt.Type == BinaryAck {
		str += fmt.Sprintf("%d-", pkt.Attachments)
	}

	if len(pkt.Nsp) > 0 && pkt.Nsp != "/" {
		str += fmt.Sprintf("%s,", pkt.Nsp)
	}

	if pkt.ID != nil {
		str += strconv.FormatUint(*pkt.ID, 10)
	}

	if pkt.Data != nil {
		if tmp, ok := pkt.Data.(string); ok {
			str += tmp
		} else {
			tmp, err := json.Marshal(pkt.Data)
			if err == nil {
				str += string(tmp)
			}
		}
	}

	return str
}

func (e *encoder) hasBinary(data any) bool {
	if data == nil {
		return false
	}

	switch data.(type) {
	case []byte, [][]byte, map[string][]byte:
		return true
	}

	return false
}

func (d *deconstructor) do(pkt *Packet) []any {
	d.buffers = make([]any, 0)

	pkt.Data = d.deconstruct(pkt.Data)
	pkt.Attachments = len(d.buffers)

	return d.buffers
}

func (d *deconstructor) deconstruct(data any) any {
	if data == nil {
		return nil
	}

	switch tmp := data.(type) {
	case []byte:
		d.buffers = append(d.buffers, tmp)
		return map[string]any{
			KeyPlaceholder: true,
			KeyNum:         len(d.buffers) - 1,
		}
	case [][]byte:
		newData := make([]any, 0, len(tmp))
		for _, item := range tmp {
			newData = append(newData, d.deconstruct(item))
		}
		return newData
	case map[string][]byte:
		newData := make(map[string]any, len(tmp))
		for key, item := range tmp {
			newData[key] = d.deconstruct(item)
		}
		return newData
	}

	return data
}

func Decode(callback Callback, data ...any) error {
	d := NewDecoder(callback)
	for _, item := range data {
		if err := d.Add(item); err != nil {
			return err
		}
	}
	return nil
}

func NewDecoder(callback Callback) *decoder {
	return &decoder{callback: callback}
}

func (d *decoder) Add(data any) error {
	switch tmp := data.(type) {
	case string:
		pkt, err := d.decodeString(tmp)
		if err != nil {
			return err
		}
		if pkt.Type == BinaryEvent || pkt.Type == BinaryAck {
			d.reconstructor = newReconstructor(&pkt)

			if pkt.Attachments != 0 {
				return nil
			}
		}

		if d.callback != nil {
			return d.callback(pkt)
		}
	case []byte:
		if d.reconstructor == nil {
			return ErrReconstructPacket
		}
		if pkt := d.reconstructor.takeBinaryData(tmp); pkt != nil {
			d.reconstructor = nil
			if d.callback != nil {
				return d.callback(*pkt)
			}
		}
	}
	return nil
}

func (d *decoder) decodeString(str string) (Packet, error) {
	var pkt Packet

	if len(str) == 0 {
		return pkt, ErrEmptyData
	}

	t, err := ParseTypeASCII(str[0])
	if err != nil {
		return pkt, err
	}

	str = str[1:]
	pkt.Type = t

	if BinaryEvent == t || BinaryAck == t {
		b, a, _ := strings.Cut(str, "-")
		str = a
		pkt.Attachments, err = strconv.Atoi(b)
		if err != nil {
			return pkt, ErrIllegalAttachments
		}
	}

	if len(str) > 0 && '/' == str[0] {
		pkt.Nsp, str, _ = strings.Cut(str, ",")
	} else {
		pkt.Nsp = "/"
	}

	if len(str) > 0 {
		var n uint64
		for i, ch := range []byte(str) {
			ch -= '0'
			if ch > 9 {
				if i > 0 {
					str = str[i:]
					pkt.ID = &n
				}
				break
			}
			n = n*10 + uint64(ch)
		}
	}

	if len(str) > 0 {
		if Disconnect == t {
			return pkt, ErrInvalidPayload
		}
		pkt.Data = str
	}

	return pkt, nil
}

func (d *decoder) Destroy() {
	if d.reconstructor != nil {
		d.reconstructor.finishReconstruction()
	}
	d.callback = nil
}

func (d *decoder) OnDecoded(callback Callback) {
	d.callback = callback
}

func newReconstructor(pkt *Packet) *reconstructor {
	return &reconstructor{
		pkt:     pkt,
		buffers: make([]any, 0),
	}
}

func (r *reconstructor) takeBinaryData(data []byte) *Packet {
	r.buffers = append(r.buffers, data)

	if len(r.buffers) == r.pkt.Attachments {
		r.reconstruct()
		r.finishReconstruction()
		return r.pkt
	}
	return nil
}

func (r *reconstructor) reconstruct() {
	r.pkt.Data = r.reconstructData(r.pkt.Data)
	r.pkt.Attachments = -1
}

func (r *reconstructor) reconstructData(data any) any {
	// TODO
	return r.buffers

	//return data
}

func (r *reconstructor) finishReconstruction() {
	r.pkt = nil
	r.buffers = nil
}
