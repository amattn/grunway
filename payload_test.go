package grunway

import (
	"testing"
)

func TestUnmarshallBad(t *testing.T) {
	var err error
	_, err = UnmarshalPayloadWrapper([]byte{})
	if err == nil {
		t.Error("UnmarshalPayloadWrapper should error on empty bytes input")
	}
	_, err = UnmarshalPayloadWrapper([]byte("xxx"), AuthorPayload{})
	if err == nil {
		t.Error("UnmarshalPayloadWrapper should error on non-json input")
	}
	_, err = UnmarshalPayloadWrapper([]byte("{}"))
	if err == nil {
		t.Error("UnmarshalPayloadWrapper should error when len(supportedPayloads) == 0")
	}
}

func TestUnmarshall(t *testing.T) {
	auth1 := AuthorPayload{}
	book1 := BookPayload{}
	book2 := BookPayload{}

	auth1.Name = "Me"
	auth1.PKey = 1
	book1.AuthorId = 1
	book1.Name = "Marshalling for Dummies"
	book2.AuthorId = 1
	book2.Name = "Unmarshalling for Dummies"

	pw := NewPayloadWrapper(auth1, book1, book2)
	jsonBytes, jsonErr := MarshallPayloadWrapper(pw)
	if jsonErr != nil {
		t.Logf("92740011906 pw:%+v", pw)
		t.Fatal("92740011906", "MarshallPayloadWrapper failure", jsonErr)
	}

	upw, err := UnmarshalPayloadWrapper(jsonBytes, AuthorPayload{}, BookPayload{})
	if err != nil {
		t.Logf("92740011907 jsonBytes:%+v", string(jsonBytes))
		t.Error("92740011907", "UnmarshalPayloadWrapper failure", err)
	}

	if len(upw.Alert) != 0 {
		t.Logf("92710202703 upw:%+v", upw)
		t.Error("92710202703", "UnmarshalPayloadWrapper failure len(upw.Alert) != 0", err)
	}
	if len(upw.ErrorMessage) != 0 {
		t.Logf("92710202704 upw:%+v", upw)
		t.Error("92710202704", "UnmarshalPayloadWrapper failure len(upw.ErrorMessage) != 0", err)
	}
	if upw.ErrorNumber != 0 {
		t.Logf("92710202705 upw:%+v", upw)
		t.Error("92710202705", "UnmarshalPayloadWrapper failure upw.ErrorNumber != 0", err)
	}

	if len(upw.Payloads) != 2 {
		t.Logf("92710202706 upw:%+v", upw)
		t.Fatal("92710202706", "UnmarshalPayloadWrapper failure upw.ErrorNumber != 0", err)
	}

	for payloadType, payloadList := range upw.Payloads {
		if payloadType == "book" && len(payloadList) != 2 {
			t.Errorf("91759077316 Incorrect number of %s payloads:%+v\n%+v", payloadType, payloadList, upw)
		}

		if payloadType == "author" {
			if len(payloadList) != 1 {
				t.Errorf("91759077316 Incorrect number of %s payloads:%+v\n%+v", payloadType, payloadList, upw)
			} else {
				payload := payloadList[0]
				author, ok := payload.(*AuthorPayload)
				if ok {
					if author.Name != "Me" {
						t.Errorf("91865607778 expected author.Name == Me, got %v\n%+v", author.Name, author)
					}
				} else {
					t.Errorf("91865607779 expected *AuthorPayload, got %T\n%+v", author, payload)
				}
			}
		}
	}
}
