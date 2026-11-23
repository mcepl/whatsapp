package connector

import (
	"testing"

	"go.mau.fi/whatsmeow/types"
	"maunium.net/go/mautrix/bridgev2/database"
)

func TestDisplaynameQuality(t *testing.T) {
	templates := map[string]string{
		"default":      `{{or .BusinessName .PushName .Phone .RedactedPhone "Unknown user"}} (WA)`,
		"contact name": `{{or .FullName .BusinessName .PushName .Phone .RedactedPhone "Unknown user"}} (WA)`,
		// Fails without a phone number, which the startup check doesn't catch
		"slice phone": `{{or .PushName (slice .Phone 1)}}`,
		"found":       `{{if .Found}}{{or .PushName .Phone "Unknown"}}{{else}}Unknown{{end}}`,
	}
	pnJID := types.NewJID("12345678900", types.DefaultUserServer)
	lidJID := types.NewJID("987654321", types.HiddenUserServer)
	tests := []struct {
		template string
		jid      types.JID
		phone    string
		contact  types.ContactInfo
		want     database.NameQuality
	}{
		{"default", pnJID, "", types.ContactInfo{}, database.NameQualityIdentifier},
		{"default", lidJID, "", types.ContactInfo{RedactedPhone: "+1∙∙∙∙∙∙∙∙00"}, database.NameQualityIdentifier},
		{"default", lidJID, "", types.ContactInfo{}, database.NameQualityIdentifier},
		{"default", pnJID, "", types.ContactInfo{PushName: "Alice"}, database.NameQualityProfile},
		{"default", lidJID, "+12345678900", types.ContactInfo{BusinessName: "Alice Ltd"}, database.NameQualityProfile},
		// The default template doesn't use contact names, so they don't raise the quality
		{"default", pnJID, "", types.ContactInfo{FullName: "Ali", PushName: "Alice"}, database.NameQualityProfile},
		{"contact name", pnJID, "", types.ContactInfo{FullName: "Ali", PushName: "Alice"}, database.NameQualityContact},
		{"contact name", pnJID, "", types.ContactInfo{PushName: "Alice"}, database.NameQualityProfile},
		{"contact name", pnJID, "", types.ContactInfo{}, database.NameQualityIdentifier},
		{"slice phone", lidJID, "", types.ContactInfo{PushName: "Alice"}, database.NameQualityUnranked},
		{"slice phone", pnJID, "", types.ContactInfo{PushName: "Alice"}, database.NameQualityProfile},
		{"found", pnJID, "", types.ContactInfo{Found: true}, database.NameQualityIdentifier},
		{"found", pnJID, "", types.ContactInfo{Found: true, PushName: "Alice"}, database.NameQualityProfile},
	}
	for _, tt := range tests {
		cfg := &Config{DisplaynameTemplate: templates[tt.template]}
		if err := cfg.PostProcess(); err != nil {
			t.Fatalf("failed to parse template %q: %v", tt.template, err)
		}
		name := cfg.FormatDisplayname(tt.jid, tt.phone, tt.contact)
		if got := cfg.displaynameQuality(tt.jid, tt.phone, tt.contact, name); got != tt.want {
			t.Errorf("template %q, contact %+v: name %q got quality %d, want %d", tt.template, tt.contact, name, got, tt.want)
		}
	}
}
