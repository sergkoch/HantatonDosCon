<script lang="ts">
  import { onMount } from "svelte";
  import { t } from "../lib/i18n.svelte";

  const KEY = "cookie-consent";
  let visible = $state(false);

  onMount(() => {
    try {
      visible = localStorage.getItem(KEY) !== "accepted";
    } catch {
      visible = true;
    }
  });

  function accept() {
    try {
      localStorage.setItem(KEY, "accepted");
    } catch {
      // localStorage недоступен — просто скрываем баннер
    }
    visible = false;
  }
</script>

{#if visible}
  <div
    class="cookie-consent"
    role="region"
    aria-label={t("cookie.label")}
  >
    <p class="cookie-consent__text">
      {t("cookie.text")}
    </p>
    <button class="button button--primary" type="button" onclick={accept}>
      {t("cookie.accept")}
    </button>
  </div>
{/if}
