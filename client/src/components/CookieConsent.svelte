<script lang="ts">
  import { onMount } from "svelte";
  import {
    acceptConsent,
    closeConsent,
    consentOpen,
    initConsent,
  } from "../lib/cookieConsent.svelte";
  import { t } from "../lib/i18n.svelte";

  let detailsOpen = $state(false);

  onMount(() => {
    initConsent();
  });

  const open = $derived(consentOpen());
</script>

{#if open}
  <div class="cookie-consent" role="region" aria-label={t("cookie.label")}>
    <div class="cookie-consent__head">
      <p class="cookie-consent__title">{t("cookie.label")}</p>
      <button
        class="cookie-consent__close"
        type="button"
        onclick={closeConsent}
        aria-label={t("cookie.close")}
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="m6 6 12 12M18 6 6 18"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
      </button>
    </div>

    <p class="cookie-consent__text">{t("cookie.text")}</p>
    <p class="cookie-consent__text cookie-consent__text--muted">
      {t("cookie.noAnalytics")}
    </p>

    <button
      class="cookie-consent__what"
      type="button"
      aria-expanded={detailsOpen}
      aria-controls="cookie-details"
      onclick={() => (detailsOpen = !detailsOpen)}
    >
      {t("cookie.what")}
      <svg
        class={"cookie-consent__chevron" + (detailsOpen ? " is-open" : "")}
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
      >
        <path
          d="m6 9 6 6 6-6"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </button>

    {#if detailsOpen}
      <ul class="cookie-consent__keys" id="cookie-details">
        <li><code>theme</code> — {t("cookie.keyTheme")}</li>
        <li><code>locale</code> — {t("cookie.keyLocale")}</li>
        <li><code>ac.client-id</code> — {t("cookie.keyClient")}</li>
        <li><code>cookie-consent</code> — {t("cookie.keyConsent")}</li>
      </ul>
    {/if}

    <button class="button button--primary" type="button" onclick={acceptConsent}>
      {t("cookie.accept")}
    </button>
  </div>
{/if}
