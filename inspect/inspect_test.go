package inspect

import (
	"testing"
)

func TestGenerateHex(t *testing.T) {
	const hexStr = "2F3FC99486D46537130FD12E072B1F2B17A1F580C32C6FCB2E47075F2BB3D6DC85"
	item, err := DecodeHex(hexStr)
	if err != nil {
		t.Fatalf("DecodeHex failed: %v", err)
	}
	generated, err := GenerateHex(item)
	if err != nil {
		t.Fatalf("GenerateHex failed: %v", err)
	}
	t.Logf("Generated hex: %s", generated)
	decoded, err := DecodeHex(generated)
	if err != nil {
		t.Fatalf("DecodeHex of generated hex failed: %v", err)
	}
	t.Logf("Round-trip ItemPreviewData: %+v", decoded)
}

func TestGenerateLink(t *testing.T) {
	const hexStr = "2F3FC99486D46537130FD12E072B1F2B17A1F580C32C6FCB2E47075F2BB3D6DC85"
	item, err := DecodeHex(hexStr)
	if err != nil {
		t.Fatalf("DecodeHex failed: %v", err)
	}
	link, err := GenerateLink(item)
	if err != nil {
		t.Fatalf("GenerateLink failed: %v", err)
	}
	t.Logf("Generated link: %s", link)
	decoded, err := DecodeLink(link)
	if err != nil {
		t.Fatalf("DecodeLink of generated link failed: %v", err)
	}
	t.Logf("Round-trip ItemPreviewData: %+v", decoded)
}

func TestDecodeLinkAndGenerateWithStickers(t *testing.T) {
	const link = "steam://run/730//+csgo_econ_action_preview%20A6B648287E21E1BEAF8668A38EA296A29E7C1E4354A5E6F7C4A3AEA7B61485C4A3AEA4B60085C4A3AEA5B60585CE25262626AAD6AEE8D94E90"
	item, err := DecodeLink(link)
	if err != nil {
		t.Fatalf("DecodeLink failed: %v", err)
	}
	t.Logf("Decoded ItemPreviewData: %+v", item)
	t.Logf("Stickers: %+v", item.GetStickers())
	generated, err := GenerateLink(item)
	if err != nil {
		t.Fatalf("GenerateLink failed: %v", err)
	}
	t.Logf("Generated link: %s", generated)
}

func TestDecodeLinkAndGenerate(t *testing.T) {
	const link = "steam://run/730//+csgo_econ_action_preview%202F3FC99486D46537130FD12E072B1F2B17A1F580C32C6FCB2E47075F2BB3D6DC85"
	item, err := DecodeLink(link)
	if err != nil {
		t.Fatalf("DecodeLink failed: %v", err)
	}
	t.Logf("Decoded ItemPreviewData: %+v", item)
	generated, err := GenerateLink(item)
	if err != nil {
		t.Fatalf("GenerateLink failed: %v", err)
	}
	t.Logf("Generated link: %s", generated)
}

func TestDecodeHex(t *testing.T) {
	const hexStr = "2F3FC99486D46537130FD12E072B1F2B17A1F580C32C6FCB2E47075F2BB3D6DC85"
	item, err := DecodeHex(hexStr)
	if err != nil {
		t.Fatalf("DecodeHex failed: %v", err)
	}
	t.Logf("ItemPreviewData: %+v", item)
	t.Logf("PaintWear (float): %.15f", BytesToFloat(item.GetPaintwear()))
}
