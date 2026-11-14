import { parse } from "parse5";
import { pathToFileURL } from "node:url";

const storefrontDomainCheckTimeoutMilliseconds = 15000;
const storefrontCanonicalProbeRetryDelaysMilliseconds = [500, 1500];

const pauseBeforeStorefrontCanonicalProbeRetry = (delayMilliseconds) =>
  new Promise((resolve) => setTimeout(resolve, delayMilliseconds));

export const parseStorefrontCanonicalHttpsOrigin = (configuredValue) => {
  const trimmedValue = String(configuredValue || "").trim();
  if (!trimmedValue) {
    throw new Error(
      "Set the repository Actions variable STOREFRONT_SITE_URL to the production canonical HTTPS origin.",
    );
  }

  let configuredUrl;
  try {
    configuredUrl = new URL(trimmedValue);
  } catch {
    throw new Error(
      "STOREFRONT_SITE_URL must be a valid HTTPS origin, such as https://shop.example.com.",
    );
  }

  if (
    configuredUrl.protocol !== "https:" ||
    configuredUrl.username ||
    configuredUrl.password ||
    (configuredUrl.pathname !== "/" && configuredUrl.pathname !== "") ||
    configuredUrl.search ||
    configuredUrl.hash
  ) {
    throw new Error(
      "STOREFRONT_SITE_URL must contain only the production canonical HTTPS origin, such as https://shop.example.com.",
    );
  }

  return configuredUrl.origin;
};

export const readStorefrontCanonicalOriginFromHtml = (html, pageUrl) => {
  const document = parse(String(html));
  const headElements = [];
  const pendingNodes = [document];

  while (pendingNodes.length > 0) {
    const node = pendingNodes.pop();
    if (node.tagName === "head") headElements.push(node);
    if (Array.isArray(node.childNodes)) pendingNodes.push(...node.childNodes);
  }

  if (headElements.length === 0) {
    throw new Error(
      "The live storefront homepage did not contain an HTML head to inspect.",
    );
  }

  const canonicalOrigins = [];
  for (const headElement of headElements) {
    const pendingHeadNodes = [...(headElement.childNodes || [])];
    while (pendingHeadNodes.length > 0) {
      const node = pendingHeadNodes.pop();
      if (Array.isArray(node.childNodes)) {
        pendingHeadNodes.push(...node.childNodes);
      }
      if (node.tagName !== "link") continue;

      const attributes = new Map(
        (node.attrs || []).map(({ name, value }) => [name, value]),
      );
      const relValues = String(attributes.get("rel") || "")
        .toLowerCase()
        .split(/\s+/);
      if (!relValues.includes("canonical")) continue;

      const href = attributes.get("href");
      if (!href) {
        throw new Error(
          "The live storefront has a canonical link without an href.",
        );
      }

      let canonicalUrl;
      try {
        canonicalUrl = new URL(href, pageUrl);
      } catch {
        throw new Error(
          `The live storefront has an invalid canonical URL: ${href}`,
        );
      }

      if (
        canonicalUrl.protocol !== "https:" ||
        canonicalUrl.username ||
        canonicalUrl.password
      ) {
        throw new Error(
          `The live storefront canonical URL is not a secure public URL: ${canonicalUrl.href}`,
        );
      }
      canonicalOrigins.push(canonicalUrl.origin);
    }
  }

  if (canonicalOrigins.length === 0) {
    throw new Error(
      "The live storefront homepage did not contain a canonical link.",
    );
  }

  const uniqueCanonicalOrigins = [...new Set(canonicalOrigins)];
  if (uniqueCanonicalOrigins.length !== 1) {
    throw new Error(
      `The live storefront homepage contains conflicting canonical origins: ${uniqueCanonicalOrigins.join(", ")}`,
    );
  }

  return uniqueCanonicalOrigins[0];
};

const fetchStorefrontHomepageForCanonicalProbe = async ({
  configuredOrigin,
  fetchImplementation,
  pauseImplementation,
}) => {
  let lastNetworkError = "";

  for (
    let attempt = 0;
    attempt <= storefrontCanonicalProbeRetryDelaysMilliseconds.length;
    attempt += 1
  ) {
    try {
      const response = await fetchImplementation(`${configuredOrigin}/`, {
        redirect: "follow",
        headers: {
          accept: "text/html",
          "user-agent": "storefront-publish-domain-guard/1.0",
        },
        signal: AbortSignal.timeout(storefrontDomainCheckTimeoutMilliseconds),
      });

      const isRetryableServerError =
        response.status >= 500 && response.status < 600;
      const hasAnotherAttempt =
        attempt < storefrontCanonicalProbeRetryDelaysMilliseconds.length;
      if (isRetryableServerError) {
        if (!hasAnotherAttempt) return { response, html: "" };
        await response.body?.cancel();
        lastNetworkError = `HTTP ${response.status}`;
      } else if (!response.ok) {
        return { response, html: "" };
      } else if (
        !(response.headers?.get("content-type") || "")
          .toLowerCase()
          .includes("text/html")
      ) {
        return { response, html: "" };
      } else {
        return { response, html: await response.text() };
      }
    } catch (error) {
      lastNetworkError = error instanceof Error ? error.message : String(error);
      if (attempt === storefrontCanonicalProbeRetryDelaysMilliseconds.length) {
        throw new Error(
          `The storefront canonical check failed after ${attempt + 1} attempts: ${lastNetworkError}`,
        );
      }
    }

    await pauseImplementation(
      storefrontCanonicalProbeRetryDelaysMilliseconds[attempt],
    );
  }

  throw new Error(
    `The storefront canonical check failed after ${storefrontCanonicalProbeRetryDelaysMilliseconds.length + 1} attempts: ${lastNetworkError}`,
  );
};

export const validateStorefrontPublishDomainAgainstLiveCanonical = async ({
  configuredSiteUrl,
  allowPlannedDomainMigration = false,
  currentConfiguredSiteUrl = "",
  fetchImplementation = fetch,
  pauseImplementation = pauseBeforeStorefrontCanonicalProbeRetry,
}) => {
  const configuredOrigin =
    parseStorefrontCanonicalHttpsOrigin(configuredSiteUrl);

  if (allowPlannedDomainMigration) {
    const currentConfiguredOrigin = currentConfiguredSiteUrl
      ? parseStorefrontCanonicalHttpsOrigin(currentConfiguredSiteUrl)
      : "";
    if (currentConfiguredOrigin === configuredOrigin) {
      throw new Error(
        "The migration input matches STOREFRONT_SITE_URL. Leave it blank for a routine publish so the live canonical domain is checked.",
      );
    }

    return {
      configuredOrigin,
      liveCanonicalOrigin: null,
      migrationCheckSkipped: true,
    };
  }

  let response;
  let html;
  try {
    ({ response, html } = await fetchStorefrontHomepageForCanonicalProbe({
      configuredOrigin,
      fetchImplementation,
      pauseImplementation,
    }));
  } catch (error) {
    throw new Error(
      `Could not inspect the live storefront canonical domain at ${configuredOrigin}: ${error instanceof Error ? error.message : String(error)}. Publication is stopped; restore public homepage access or use the planned domain migration input only for an approved domain change.`,
    );
  }

  if (!response.ok) {
    throw new Error(
      `The live storefront homepage at ${configuredOrigin} returned HTTP ${response.status}; publication is stopped because its canonical domain could not be verified.`,
    );
  }

  const contentType = response.headers?.get("content-type") || "";
  if (!contentType.toLowerCase().includes("text/html")) {
    throw new Error(
      `The live storefront homepage at ${configuredOrigin} returned ${contentType || "no content type"}, not HTML; publication is stopped because its canonical domain could not be verified.`,
    );
  }

  const finalPageUrl = response.url || `${configuredOrigin}/`;
  const finalResponseOrigin = new URL(finalPageUrl).origin;
  const liveCanonicalOrigin = readStorefrontCanonicalOriginFromHtml(
    html,
    finalPageUrl,
  );
  if (
    finalResponseOrigin !== configuredOrigin ||
    liveCanonicalOrigin !== configuredOrigin
  ) {
    throw new Error(
      `STOREFRONT_SITE_URL is ${configuredOrigin}, but the live homepage resolves to ${finalResponseOrigin} and declares canonical origin ${liveCanonicalOrigin}. Update the repository Actions variable. For a planned domain migration, dispatch Publish Production Images and enter the new HTTPS origin in the storefront_site_url input; the workflow saves it after all images publish successfully.`,
    );
  }

  return {
    configuredOrigin,
    liveCanonicalOrigin,
    migrationCheckSkipped: false,
  };
};

const runStorefrontPublishDomainGuard = async () => {
  const allowPlannedDomainMigration =
    process.env.STOREFRONT_DOMAIN_MIGRATION === "true";
  const result = await validateStorefrontPublishDomainAgainstLiveCanonical({
    configuredSiteUrl: process.env.STOREFRONT_SITE_URL,
    allowPlannedDomainMigration,
    currentConfiguredSiteUrl: process.env.STOREFRONT_CURRENT_SITE_URL,
  });

  if (result.migrationCheckSkipped) {
    console.log(
      `Approved domain migration target ${result.configuredOrigin} is used for this storefront build. The repository Actions variable will be saved after every image publishes successfully.`,
    );
    return;
  }

  console.log(
    `Live storefront canonical origin matches STOREFRONT_SITE_URL (${result.configuredOrigin}).`,
  );
};

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(process.argv[1]).href
) {
  runStorefrontPublishDomainGuard().catch((error) => {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
  });
}
