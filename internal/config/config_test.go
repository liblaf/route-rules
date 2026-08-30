package config

import "testing"

func TestValidateSourceAcceptsLocalCIDRFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "sources/tailnet.ipcidr.txt"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSourceRejectsEscapingLocalFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "../tailnet.ipcidr.txt"}); err == nil {
		t.Fatal("escaping local path did not fail validation")
	}
}
