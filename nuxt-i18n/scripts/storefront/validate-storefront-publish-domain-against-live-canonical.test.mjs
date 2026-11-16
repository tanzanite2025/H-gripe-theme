import assert from "node:assert/strict";
import test from "node:test";
import {
  parseStorefrontCanonicalHttpsOrigin,
  readStorefrontCanonicalOriginFromHtml,
  validateStorefrontPublishDomainAgainstLiveCanonical,
} from "./validate-storefront-publish-domain-against-live-canonical.mjs";

const createHtmlResponse = ({
  html = '<html><head><link rel="canonical" href="https://shop.example.test/" /></head><body></body></html>',
  responseUrl = "https://shop.example.test/",
  status = 200,
  contentType = "text/html; charset=utf-8",
} = {}) => ({
  ok: status >= 200 && status < 300,
  status,
  url: responseUrl,
  headers: new Headers({ "content-type": contentType }),
  text: async () => html,
});

test("accepts and normalizes an HTTPS canonical origin", () => {
  assert.equal(
    parseStorefrontCanonicalHttpsOrigin(" HTTPS://SHOP.EXAMPLE.TEST/ "),
    "https://shop.example.test",
  );
});

test("rejects missing, insecure, credentialed, or path-based site URLs", () => {
  for (const value of [
    "",
    "http://shop.example.test",
    "https://user:password@shop.example.test",
    "https://shop.example.test/store",
    "https://shop.example.test/?campaign=1",
    "https://shop.example.test/#section",
    "not a URL",
  ]) {
    assert.throws(
      () => parseStorefrontCanonicalHttpsOrigin(value),
      value || "empty URL",
    );
  }
});

test("parses canonical link attributes and HTML entities with HTML semantics", () => {
  assert.equal(
    readStorefrontCanonicalOriginFromHtml(
      '<html><head><script>const sample = "<link rel=canonical href=https://wrong.example.test>";</script><link data-note=">" href="/shop?size=700&amp;type=road" rel="alternate canonical" /></head></html>',
      "https://shop.example.test/",
    ),
    "https://shop.example.test",
  );
});

test("accepts a live homepage whose response and canonical origins match the configured origin", async () => {
  let requestedUrl = "";
  const result = await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: "https://shop.example.test",
    fetchImplementation: async (url) => {
      requestedUrl = url;
      return createHtmlResponse();
    },
  });

  assert.equal(requestedUrl, "https://shop.example.test/");
  assert.equal(result.liveCanonicalOrigin, "https://shop.example.test");
  assert.equal(result.migrationCheckSkipped, false);
});

test("sends the dedicated publish probe header when a token is configured", async () => {
  let requestHeaders;
  const probeToken = "test-only-publish-probe-token";

  await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: "https://shop.example.test",
    publishProbeToken: probeToken,
    fetchImplementation: async (_url, options) => {
      requestHeaders = options.headers;
      return createHtmlResponse();
    },
  });

  assert.equal(
    requestHeaders["x-storefront-publish-domain-probe"],
    probeToken,
  );
});

test("retries a transient connection failure before accepting the live canonical", async () => {
  let attempts = 0;
  let pauses = 0;
  const result = await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: "https://shop.example.test",
    fetchImplementation: async () => {
      attempts += 1;
      if (attempts === 1) throw new Error("temporary connection reset");
      return createHtmlResponse();
    },
    pauseImplementation: async () => {
      pauses += 1;
    },
  });

  assert.equal(attempts, 2);
  assert.equal(pauses, 1);
  assert.equal(result.liveCanonicalOrigin, "https://shop.example.test");
});

test("retries transient server errors but does not retry a CDN denial", async () => {
  let serverErrorAttempts = 0;
  const serverErrorResult =
    await validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      fetchImplementation: async () => {
        serverErrorAttempts += 1;
        return serverErrorAttempts < 3
          ? createHtmlResponse({ status: 503 })
          : createHtmlResponse();
      },
      pauseImplementation: async () => {},
    });
  assert.equal(serverErrorAttempts, 3);
  assert.equal(
    serverErrorResult.liveCanonicalOrigin,
    "https://shop.example.test",
  );

  let denialAttempts = 0;
  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      fetchImplementation: async () => {
        denialAttempts += 1;
        return createHtmlResponse({ status: 403 });
      },
      pauseImplementation: async () => {},
    }),
    /returned HTTP 403/,
  );
  assert.equal(denialAttempts, 1);
});

test("retries when the homepage response body fails before completion", async () => {
  let attempts = 0;
  const result = await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: "https://shop.example.test",
    fetchImplementation: async () => {
      attempts += 1;
      if (attempts === 1) {
        return {
          ...createHtmlResponse(),
          text: async () => {
            throw new Error("temporary response-body interruption");
          },
        };
      }
      return createHtmlResponse();
    },
    pauseImplementation: async () => {},
  });

  assert.equal(attempts, 2);
  assert.equal(result.liveCanonicalOrigin, "https://shop.example.test");
});

test("blocks publication when the live canonical origin has changed", async () => {
  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://old.example.test",
      fetchImplementation: async () =>
        createHtmlResponse({
          responseUrl: "https://old.example.test/",
          html: '<html><head><link rel="canonical" href="https://new.example.test/" /></head></html>',
        }),
    }),
    /Update the repository Actions variable/,
  );
});

test("blocks publication when the configured hostname redirects to a replacement domain", async () => {
  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://old.example.test",
      fetchImplementation: async () =>
        createHtmlResponse({
          responseUrl: "https://new.example.test/",
          html: '<html><head><link rel="canonical" href="https://new.example.test/" /></head></html>',
        }),
    }),
    /live homepage resolves to https:\/\/new\.example\.test/,
  );
});

test("blocks ambiguous canonical metadata, non-HTML responses, and CDN errors", async () => {
  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      fetchImplementation: async () =>
        createHtmlResponse({
          html: '<html><head><link rel="canonical" href="https://shop.example.test/" /><link rel="canonical" href="https://other.example.test/" /></head></html>',
        }),
    }),
    /conflicting canonical origins/,
  );

  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      fetchImplementation: async () =>
        createHtmlResponse({ contentType: "application/json" }),
    }),
    /not HTML/,
  );

  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      fetchImplementation: async () => createHtmlResponse({ status: 403 }),
    }),
    /returned HTTP 403/,
  );
});

test("allows the explicit migration target without requiring the new domain to be live yet", async () => {
  const result = await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: "https://new.example.test",
    allowPlannedDomainMigration: true,
    currentConfiguredSiteUrl: "https://old.example.test",
    fetchImplementation: async () => {
      throw new Error(
        "A planned domain target is not expected to be live yet.",
      );
    },
  });

  assert.equal(result.configuredOrigin, "https://new.example.test");
  assert.equal(result.liveCanonicalOrigin, null);
  assert.equal(result.migrationCheckSkipped, true);
});

test("does not allow the migration input to bypass checks for the current domain", async () => {
  await assert.rejects(
    validateStorefrontPublishDomainAgainstLiveCanonical({
      configuredSiteUrl: "https://shop.example.test",
      currentConfiguredSiteUrl: "https://SHOP.example.test/",
      allowPlannedDomainMigration: true,
    }),
    /matches STOREFRONT_SITE_URL/,
  );
});
