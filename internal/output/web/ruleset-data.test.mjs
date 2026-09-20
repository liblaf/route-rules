import assert from "node:assert/strict";
import test from "node:test";

import { createRulesetStore } from "./ruleset-data.mjs";

const metadata = {
  priority: ["global"],
  fallback: "direct",
  rulesets: [{ name: "global" }],
  artifacts: [
    {
      path: "global.domain.list",
      ruleset: "global",
      behavior: "domain",
      format: "text",
      rule_count: 2,
    },
    {
      path: "global.domain.mrs",
      ruleset: "global",
      behavior: "domain",
      format: "mrs",
      rule_count: 2,
    },
    {
      path: "global.ipcidr.list",
      ruleset: "global",
      behavior: "ipcidr",
      format: "text",
      rule_count: 1,
    },
    {
      path: "global.ipcidr.mrs",
      ruleset: "global",
      behavior: "ipcidr",
      format: "mrs",
      rule_count: 1,
    },
  ],
};

function response(body, status = 200) {
  return new Response(body, {
    status,
    statusText: status === 200 ? "OK" : "Not Found",
  });
}

test("loads and caches text artifacts while skipping MRS files", async () => {
  const requests = [];
  const store = createRulesetStore({
    baseUrl: "https://rules.example/mihomo/",
    fetcher: async (url) => {
      requests.push(url);
      if (url.endsWith("stats.json")) return response(JSON.stringify(metadata));
      if (url.endsWith("global.domain.list"))
        return response("example.com\n+.example.org\n");
      if (url.endsWith("global.ipcidr.list")) return response("192.0.2.0/24\n");
      throw new Error(`unexpected request ${url}`);
    },
  });

  const [first, second] = await Promise.all([
    store.loadRules("global"),
    store.loadRules("global"),
  ]);
  assert.equal(first, second);
  assert.deepEqual(
    first.map((entry) => entry.type),
    ["DOMAIN", "DOMAIN-SUFFIX", "IP-CIDR"],
  );
  assert.ok(first.every((entry) => entry.artifact.ruleset === "global"));
  assert.deepEqual(requests.sort(), [
    "https://rules.example/mihomo/global.domain.list",
    "https://rules.example/mihomo/global.ipcidr.list",
    "https://rules.example/mihomo/stats.json",
  ]);
  assert.equal(
    store.artifactUrl(metadata.artifacts[0]),
    "https://rules.example/mihomo/global.domain.list",
  );
});

test("failed requests are evicted so a later request retries", async () => {
  let attempts = 0;
  const store = createRulesetStore({
    baseUrl: "https://rules.example/mihomo/",
    fetcher: async (url) => {
      if (!url.endsWith("stats.json")) return response("example.com\n");
      attempts += 1;
      return attempts === 1
        ? response("unavailable", 503)
        : response(
            JSON.stringify({
              ...metadata,
              artifacts: [
                {
                  path: "global.domain.list",
                  ruleset: "global",
                  behavior: "domain",
                  format: "text",
                  rule_count: 1,
                },
              ],
            }),
          );
    },
  });
  await assert.rejects(store.loadRules("global"), /HTTP 503/);
  assert.equal((await store.loadRules("global")).length, 1);
  assert.equal(attempts, 2);
});

test("rejects incomplete binary coverage and inaccurate artifact counts", async () => {
  const incomplete = createRulesetStore({
    baseUrl: "https://rules.example/mihomo/",
    fetcher: async (url) =>
      response(
        JSON.stringify({
          ...metadata,
          artifacts: [
            {
              path: "global.domain.mrs",
              ruleset: "global",
              behavior: "domain",
              format: "mrs",
              rule_count: 1,
            },
          ],
        }),
      ),
  });
  await assert.rejects(
    incomplete.loadRules("global"),
    /no text artifact for domain/,
  );

  const mismatched = createRulesetStore({
    baseUrl: "https://rules.example/mihomo/",
    fetcher: async (url) =>
      url.endsWith("stats.json")
        ? response(
            JSON.stringify({
              ...metadata,
              artifacts: [
                {
                  path: "global.domain.list",
                  ruleset: "global",
                  behavior: "domain",
                  format: "text",
                  rule_count: 2,
                },
              ],
            }),
          )
        : response("example.com\n"),
  });
  await assert.rejects(
    mismatched.loadRules("global"),
    /expected 2 rules.*found 1/,
  );
});

test("malformed metadata is rejected and can be retried", async () => {
  let attempts = 0;
  const store = createRulesetStore({
    baseUrl: "https://rules.example/mihomo/",
    fetcher: async () =>
      response(JSON.stringify(++attempts === 1 ? { artifacts: [] } : metadata)),
  });
  await assert.rejects(store.loadMeta(), /ruleset priority and fallback/);
  assert.deepEqual(await store.loadMeta(), metadata);
});
