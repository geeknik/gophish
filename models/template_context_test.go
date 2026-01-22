package models

import (
	check "gopkg.in/check.v1"
)

type mockTemplateContext struct {
	URL           string
	FromAddress   string
	EncryptionKey string
}

func (m mockTemplateContext) getFromAddress() string {
	return m.FromAddress
}

func (m mockTemplateContext) getBaseURL() string {
	return m.URL
}

func (m mockTemplateContext) getEncryptionKey() string {
	return m.EncryptionKey
}

func (s *ModelsSuite) TestNewTemplateContext(c *check.C) {
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "Foo",
			LastName:  "Bar",
			Email:     "foo@bar.com",
		},
		RId: "1234567",
	}
	ctx := mockTemplateContext{
		URL:           "http://example.com",
		FromAddress:   "From Address <from@example.com>",
		EncryptionKey: "",
	}
	got, err := NewPhishingTemplateContext(ctx, r.BaseRecipient, r.RId)
	c.Assert(err, check.Equals, nil)
	c.Assert(got.BaseURL, check.Equals, ctx.URL)
	c.Assert(got.From, check.Equals, "From Address")
	c.Assert(got.RId, check.Equals, r.RId)
	c.Assert(got.BaseRecipient, check.DeepEquals, r.BaseRecipient)
	c.Assert(len(got.URL) > 0, check.Equals, true)
	c.Assert(len(got.TrackingURL) > 0, check.Equals, true)
	c.Assert(len(got.Tracker) > 0, check.Equals, true)
}
