import assert from "node:assert/strict";
import test from "node:test";

import { matchRules, parseQuery, parseRules } from "./ruleset-matcher.mjs";

test("domain rules preserve exact, suffix, and whole-label wildcard semantics", () => {
  const rules = parseRules("example.com\n+.example.org\n*.example.net.", "domain");
  assert.equal(matchRules(rules, parseQuery("example.com")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("api.example.org")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("EXAMPLE.ORG.")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("notexample.org")).matches.length, 0);
  assert.equal(matchRules(rules, parseQuery("api.example.com")).matches.length, 0);
  assert.equal(matchRules(rules, parseQuery("a.b.example.net")).matches.length, 0);
  assert.equal(matchRules(rules, parseQuery("a.example.net")).matches.length, 1);
  assert.equal(
    parseRules("*.example..com\n.*.example.com\n*.bad domain.com", "domain").every(
      (entry) => !entry.supported,
    ),
    true,
  );
});

test("classical rules match keyword, glob, and JS-compatible Go regex", () => {
  const rules = parseRules(
    "DOMAIN-KEYWORD,video-\nDOMAIN-WILDCARD,*.qhimgs?.com\nDOMAIN-REGEX,^api[0-9]+\\.example\\.com$\nDOMAIN-REGEX,^api[0-9]{1,3}\\.comma\\.example$",
    "classical",
  );
  assert.equal(matchRules(rules, parseQuery("cdn.qhimgs7.com")).matches[0].type, "DOMAIN-WILDCARD");
  assert.equal(
    matchRules(rules, parseQuery("video-api.invalid")).matches[0].type,
    "DOMAIN-KEYWORD",
  );
  assert.equal(matchRules(rules, parseQuery("api42.example.com")).matches[0].type, "DOMAIN-REGEX");
  assert.equal(
    matchRules(rules, parseQuery("api123.comma.example")).matches[0].type,
    "DOMAIN-REGEX",
  );
  assert.equal(
    matchRules(
      parseRules("DOMAIN-WILDCARD,*.example.com", "classical"),
      parseQuery("a.b.example.com"),
    ).matches.length,
    1,
  );
});

test("CIDR matching supports IPv4, IPv6, boundaries, and mapped IPv4", () => {
  const rules = parseRules(
    "0.0.0.0/0\n192.168.1.0/24\n192.168.1.45/32\n2001:db8::/32\n2001:db8::1/128\n::ffff:192.0.2.0/120",
    "ipcidr",
  );
  assert.equal(matchRules(rules, parseQuery("192.168.1.45")).matches.length, 3);
  assert.equal(matchRules(rules, parseQuery("192.168.2.1")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("2001:db8:1::5")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("2001:db9::1")).matches.length, 0);
  assert.equal(matchRules(rules, parseQuery("2001:db8::1")).matches.length, 2);
  assert.equal(matchRules(rules, parseQuery("::ffff:192.0.2.42")).matches.length, 1);
  assert.equal(matchRules(rules, parseQuery("2001:db8:0:0:0:0:0:1")).matches.length, 2);
  assert.equal(
    matchRules(parseRules("IP-CIDR,2001:db8::/32", "classical"), parseQuery("2001:db8::1")).matches
      .length,
    1,
  );
  assert.equal(parseQuery("2001:db8:0:0:1:0:0:1").value, "2001:db8::1:0:0:1");
  assert.equal(parseQuery("::").value, "::");
  assert.equal(parseQuery("::1").value, "::1");
  assert.equal(parseQuery("1::").value, "1::");
  assert.equal(parseQuery("0:0:0:0:0:0:0:1").value, "::1");
  assert.equal(parseQuery("2001:DB8::1").value, "2001:db8::1");
  assert.throws(() => parseQuery("2001:db8::1%eth0"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("::ffff:001.2.3.4"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("1:2:3:4:5:6:7::8"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("192.0.2.1:80"), /valid domain name or IP address/);
});

test("classical rules preserve full payloads and treat src CIDR rules as contextual", () => {
  const rules = parseRules(
    [
      "IP-CIDR,192.0.2.0/24,no-resolve",
      "IP-CIDR,198.51.100.0/24,no-resolve,src",
      "IP-CIDR6,2001:db8::/32,src",
      "AND,((NETWORK,UDP),(DST-PORT,7844))",
      "IP-ASN,13335,no-resolve,src",
    ].join("\n"),
    "classical",
  );
  assert.deepEqual(
    rules.map(({ type, value }) => [type, value]),
    [
      ["IP-CIDR", "192.0.2.0/24,no-resolve"],
      ["IP-CIDR", "198.51.100.0/24,no-resolve,src"],
      ["IP-CIDR6", "2001:db8::/32,src"],
      ["AND", "((NETWORK,UDP),(DST-PORT,7844))"],
      ["IP-ASN", "13335,no-resolve,src"],
    ],
  );
  const destMatch = matchRules(rules, parseQuery("192.0.2.10"));
  assert.equal(destMatch.matches.length, 1);
  assert.equal(destMatch.matches[0].value, "192.0.2.0/24,no-resolve");
  assert.deepEqual(
    destMatch.unsupported.map(({ type }) => type),
    ["IP-CIDR", "IP-CIDR6", "AND", "IP-ASN"],
  );
  const srcIpCheck = matchRules(rules, parseQuery("198.51.100.10"));
  assert.equal(srcIpCheck.matches.length, 0);
  assert.equal(srcIpCheck.unsupported.length, 4);
  const domainCheck = matchRules(rules, parseQuery("example.com"));
  assert.equal(domainCheck.matches.length, 0);
  assert.equal(domainCheck.unsupported.length, 4);
});

test("unsafe regexes and contextual rules are explicitly disclosed", () => {
  const result = matchRules(
    parseRules(
      "DOMAIN-REGEX,(?i)^example\\.com$\nPROCESS-NAME,curl\nGEOIP,CN\nDST-PORT,7844",
      "classical",
    ),
    parseQuery("example.com"),
  );
  assert.deepEqual(
    result.unsupported.map(({ type }) => type),
    ["DOMAIN-REGEX", "PROCESS-NAME", "GEOIP", "DST-PORT"],
  );
  const portRules = parseRules("DST-PORT,7844", "classical");
  for (const input of ["region1.v2.argotunnel.com", "198.41.192.7"]) {
    const checked = matchRules(portRules, parseQuery(input));
    assert.equal(checked.matches.length, 0);
    assert.equal(checked.unsupported[0].type, "DST-PORT");
  }
  assert.throws(() => parseQuery("https://example.com/path"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("example.com#fragment"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("127.1"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("001.2.3.4"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("0x7f000001"), /valid domain name or IP address/);
  assert.throws(() => parseQuery("2001:::1"), /valid domain name or IP address/);
  for (const value of [
    "",
    "999.1.1.1",
    "127.0.0.1:80",
    "0x7f.1",
    "%65xample.com",
    "example..com",
  ]) {
    assert.throws(() => parseQuery(value), /valid domain name or IP address/, value);
  }
});

test("unsupported entries only affect their applicable query kind", () => {
  const rules = parseRules("not a domain\n", "domain");
  assert.equal(matchRules(rules, parseQuery("192.0.2.1")).unsupported.length, 0);
  assert.equal(matchRules(rules, parseQuery("example.com")).unsupported.length, 1);
  assert.equal(parseRules("DOMAIN-REGEX,\\Qexample\\E", "classical")[0].supported, false);
  assert.equal(
    parseRules("DOMAIN-KEYWORD,\nDOMAIN-REGEX,\n", "classical").every((e) => !e.supported),
    true,
  );
});
