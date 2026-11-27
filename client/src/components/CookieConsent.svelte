<script lang="ts">
  import { onMount } from "svelte";
  import { t } from "../lib/i18n.svelte";
  import {
    acceptCookieConsent,
    closeCookieConsent,
    initCookieConsent,
    isCookieBannerOpen,
  } from "../lib/cookieConsent.svelte";

  const visible = $derived(isCookieBannerOpen());
  let detailsOpen = $state(false);

  onMount(() => {
    initCookieConsent();
  });

  const storedKeys = [
    "cookie.keyTheme",
    "cookie.keyLocale",
    "cookie.keyClient",
    "cookie.keyConsent",
  ] as const;
</script>

{#if visible}
  <div class="cookie-consent" role="region" aria-label={t("cookie.label")}>
    <div class="cookie-consent__head">
      <p class="cookie-consent__title">{t("cookie.title")}</p>
      <button
        class="cookie-consent__close"
        type="button"
        aria-label={t("cookie.close")}
        onclick={closeCookieConsent}
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
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

    <button
      class="cookie-consent__toggle"
      type="button"
      aria-expanded={detailsOpen}
      aria-controls="cookie-details"
      onclick={() => (detailsOpen = !detailsOpen)}
    >
      {detailsOpen ? t("cookie.hide") : t("cookie.more")}
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
      <div class="cookie-consent__details" id="cookie-details">
        <p class="cookie-consent__keys-title">{t("cookie.keysTitle")}</p>
        <ul class="cookie-consent__keys">
          {#each storedKeys as key (key)}
            <li>{t(key)}</li>
          {/each}
        </ul>
        <p class="cookie-consent__note">{t("cookie.noTracking")}</p>
      </div>
    {/if}

    <div class="cookie-consent__actions">
      <button class="button button--primary" type="button" onclick={acceptCookieConsent}>
        {t("cookie.accept")}
      </button>
    </div>
  </div>
{/if}
