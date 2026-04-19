package inspect

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"net/url"
	"regexp"
	"strings"

	pb "github.com/imorugiy/go-steam/csgo/protocol/protobuf"
	"google.golang.org/protobuf/proto"
)

var inspectLinkRegex = regexp.MustCompile(
	`(?i)^(?:steam://(?:run|rungame)/730/(?:76561202255233023/)?/?)?\+?csgo_econ_action_preview\s+([0-9A-F]+)$`,
)

func assertValidHex(h string) error {
	if len(h)%2 != 0 {
		return errors.New("invalid inspect hex payload: odd length")
	}
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid inspect hex payload: illegal character %q", c)
		}
	}
	return nil
}

func BytesToFloat(uintValue uint32) float32 {
	return math.Float32frombits(uintValue)
}

func FloatToBytes(f float32) uint32 {
	return math.Float32bits(f)
}

func GenerateHex(econ *pb.CEconItemPreviewDataBlock) (string, error) {
	if econ.Paintwear == nil {
		defaultPw := FloatToBytes(0.001)
		econ.Paintwear = &defaultPw
	}

	payload, err := proto.Marshal(econ)
	if err != nil {
		return "", fmt.Errorf("failed to marshal econ item: %w", err)
	}

	checksum := getChecksum(payload)
	crcBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBuf, checksum)

	buf := append([]byte{0}, payload...)
	buf = append(buf, crcBuf...)

	return strings.ToUpper(hex.EncodeToString(buf)), nil
}

func getChecksum(payload []byte) uint32 {
	data := append([]byte{0}, payload...)
	crc := crc32.ChecksumIEEE(data)

	// Mirror the JS: (crc & 0xffff) ^ (len * crc), masked to uint32
	xCRC := (crc & 0xffff) ^ (uint32(len(payload)) * crc)
	return xCRC & 0xffffffff
}

func xorMaskBuffer(buf []byte, key byte) []byte {
	out := make([]byte, len(buf))
	for i, b := range buf {
		out[i] = b ^ key
	}
	return out
}

func hasDecodedInspectPayload(e *pb.CEconItemPreviewDataBlock) bool {
	return e.Itemid != nil ||
		e.Defindex != nil ||
		e.Paintindex != nil ||
		e.Paintseed != nil ||
		len(e.Stickers) > 0 ||
		len(e.Keychains) > 0 ||
		len(e.Variations) > 0
}

func isDecodedMaskedInspectPayload(e *pb.CEconItemPreviewDataBlock) bool {
	return e.Itemid != nil &&
		e.Defindex != nil &&
		e.Paintindex != nil &&
		e.Inventory != nil &&
		e.Origin != nil
}

func fromBinary(payload []byte) (*pb.CEconItemPreviewDataBlock, error) {
	msg := &pb.CEconItemPreviewDataBlock{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func decodeURIComponentSafely(value string) string {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}

func extractHexFromLink(link string) (string, error) {
	decodedLink := decodeURIComponentSafely(strings.TrimSpace(link))
	matches := inspectLinkRegex.FindStringSubmatch(decodedLink)
	if len(matches) < 2 {
		return "", errors.New("invalid inspect link")
	}
	hexStr := matches[1]
	if err := assertValidHex(hexStr); err != nil {
		return "", err
	}
	return hexStr, nil
}

func decodePayload(payload []byte, validate func(*pb.CEconItemPreviewDataBlock) bool) (*pb.CEconItemPreviewDataBlock, error) {
	econ, err := fromBinary(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid inspect hex payload: %w", err)
	}
	if !validate(econ) {
		return nil, errors.New("invalid inspect hex payload")
	}
	return econ, nil
}

func decodeWrappedBuffer(buf []byte) (*pb.CEconItemPreviewDataBlock, error) {
	if len(buf) < 5 || buf[0] != 0 {
		return nil, errors.New("invalid inspect hex payload")
	}
	payload := buf[1 : len(buf)-4]
	expectedChecksum := binary.BigEndian.Uint32(buf[len(buf)-4:])
	if getChecksum(payload) != expectedChecksum {
		return nil, errors.New("inspect hex checksum missmatch")
	}
	return decodePayload(payload, hasDecodedInspectPayload)
}

func decodeMaskedBuffer(buf []byte) (*pb.CEconItemPreviewDataBlock, error) {
	if len(buf) < 5 {
		return nil, errors.New("invalid inspect hex payload")
	}
	unmasked := xorMaskBuffer(buf, buf[0])
	if unmasked[0] != 0 {
		return nil, errors.New("invalid inspect hex payload")
	}
	payload := unmasked[1 : len(unmasked)-4]
	return decodePayload(payload, isDecodedMaskedInspectPayload)
}

func DecodeHex(hexStr string) (*pb.CEconItemPreviewDataBlock, error) {
	normalized := strings.TrimSpace(hexStr)
	if err := assertValidHex(normalized); err != nil {
		return nil, err
	}
	buf, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf("invalid inspect hex payload: %w", err)
	}
	if buf[0] == 0 {
		return decodeWrappedBuffer(buf)
	}
	return decodeMaskedBuffer(buf)
}

const previewLink = "steam://rungame/730/76561202255233023/+csgo_econ_action_preview"

func GenerateLink(econ *pb.CEconItemPreviewDataBlock) (string, error) {
	hexStr, err := GenerateHex(econ)
	if err != nil {
		return "", err
	}
	return previewLink + " " + hexStr, nil
}

func DecodeLink(link string) (*pb.CEconItemPreviewDataBlock, error) {
	hexStr, err := extractHexFromLink(link)
	if err != nil {
		return nil, err
	}
	return DecodeHex(hexStr)
}
